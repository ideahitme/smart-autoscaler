package sorter

import (
	"errors"
	"fmt"
	"sync"

	"github.com/ideahitme/cluster-autoscaler/expander"
	"github.com/ideahitme/cluster-autoscaler/logger"
	"go.uber.org/zap"
)

const (
	numExpandRetries = 3
	reqBufferSize    = 1000
)

// Sorter - is a controller responsible for providing distribution of bin-items distribution
type Sorter struct {
	RequestChan chan *Request
	Bins        []*Bin
	Expander    *expander.Expander
	sync.Mutex
}

// RequestType type of the request sent to the sorter
type RequestType uint8

const (
	// RequestNewItem is sent when new item needs to be created
	RequestNewItem RequestType = iota
	// RequestDeleteItem is sent when item needs to be deleted
	RequestDeleteItem
)

// Request - is a request send to sorter
type Request struct {
	ID   string
	Type RequestType
	Item *Item
}

func (r *Request) String() (output string) {
	switch r.Type {
	case RequestNewItem:
		output += "[new item needs to be added]"
	case RequestDeleteItem:
		output += "[item needs to be deleted]"
	}
	output += fmt.Sprintf("request id:%s", r.ID)
	return
}

// Build builds the datastructure for smart decision making
func Build(bins []*Bin) *Sorter {
	return &Sorter{
		RequestChan: make(chan *Request, reqBufferSize),
	}
}

// Run runs the main controlling loop
func (s *Sorter) Run(stopChan <-chan struct{}) {
	for {
		select {
		case req := <-s.RequestChan:
			logger.Log.Info("new request", zap.String("@req_descr", req.String()))
			switch req.Type {
			case RequestNewItem:
				s.HandleNewItem(req.Item)
			case RequestDeleteItem:
				s.HandleDeleteItem(req.Item)
			}
		case <-stopChan:
			logger.Log.Info("sorter terminating")
			return
		}
	}
}

// Resync resynces the state
func (s *Sorter) Resync(bins []*Bin) {
	s.Lock()
	defer s.Unlock()

	s.Bins = bins
}

// HandleNewItem is a handler when new item is added
func (s *Sorter) HandleNewItem(item *Item) {
	s.Lock()
	defer s.Unlock()

	//determine if there is a space for the item
	if item.Ghost {
		s.tryAccomodateGhostItem(item)
		return
	}
	if !s.Expander.IsPodGhocked(item.ID) {
		if bin, err := s.tryAddBins(); err == nil {
			s.Bins = append(s.Bins, bin)
			if err := s.tryPlaceItem(bin, item); err == nil {
				bin.Items = append(bin.Items, item)
				bin.CPU -= item.CPU
				bin.Memory -= item.Memory
				return
			}
		}
	} else {
		for _, bin := range s.Bins {
			for i, binItem := range bin.Items {
				if binItem.Ghost && item.GroupID == binItem.GroupID {
					//evacuate the ghost item and place this item in its place
					if err := s.tryEvacuateItem(binItem); err == nil {
						//TODO:taint the node first
						if err := s.tryPlaceItem(bin, item); err == nil {
							if err := s.tryAccomodateGhostItem(binItem); err != nil {
								bin.Items[i] = item
								return
							}
						}
					}
				}
			}
		}
	}
	logger.Log.Error("failed to accomodate a real pod, requeuing", zap.String("@item_group_id", item.GroupID))
	s.RequestChan <- &Request{
		Type: RequestNewItem,
		ID:   newRequestID(),
		Item: item,
	}
}

func (s *Sorter) tryAccomodateGhostItem(item *Item) (finalErr error) {
	for _, bin := range s.Bins {
		if bin.CPU >= item.CPU && bin.Memory >= item.Memory {
			if err := s.tryPlaceItem(bin, item); err != nil {
				logger.Log.Error("failed to place ghost item on node", zap.String("@item_group_id", item.GroupID), zap.String("@node_id", bin.ID))
				continue
			} else {
				return
			}
		}
	}
	//no one can accomodate poor ghost item, let add another bin
	if bin, err := s.tryAddBins(); err == nil {
		s.Bins = append(s.Bins, bin)
		if err := s.tryPlaceItem(bin, item); err == nil {
			bin.CPU -= item.CPU
			bin.Memory -= item.Memory
			bin.Items = append(bin.Items, item)
			return
		}
	}
	logger.Log.Error("failed to accomodate a ghost pod, requeuing", zap.String("@item_group_id", item.GroupID))
	s.RequestChan <- &Request{
		Type: RequestNewItem,
		ID:   newRequestID(),
		Item: item,
	}
	return errors.New("accomodation has failed and therefore requeued")
}

func (s *Sorter) tryPlaceItem(bin *Bin, item *Item) error {
	for numTries := 0; numTries < numExpandRetries; numTries++ {
		if err := s.Expander.PlacePod(item.CPU, item.Memory, item.GroupID, bin.ID); err != nil {
			bin.CPU -= item.CPU
			bin.Memory -= item.Memory
			bin.Items = append(bin.Items, item)
			return nil
		}
	}
	return errors.New("failed to place ghost pod")
}

func (s *Sorter) tryAddBins() (*Bin, error) {
	for numTries := 0; numTries < numExpandRetries; numTries++ {
		if id, cpu, memory, err := s.Expander.NewNode(); err != nil {
			return &Bin{
				ID:     id,
				CPU:    cpu,
				Memory: memory,
				Items:  []*Item{},
			}, nil
		}
	}
	return nil, errors.New("failed to expand nodes")
}

func (s *Sorter) tryEvacuateItem(item *Item) error {
	for numTries := 0; numTries < numExpandRetries; numTries++ {
		if err := s.Expander.DeletePod(item.ID); err != nil {
			return nil
		}
	}
	return errors.New("failed to evacuate the item")
}

// HandleDeleteItem is a handler when item is supposed to be removed
func (s *Sorter) HandleDeleteItem(item *Item) {
}
