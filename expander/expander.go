package expander

// Expander is responsible for managing pods and scaling the cluster
type Expander struct {
}

// PlaceGhost places a ghost on a given node
func (e *Expander) PlacePod(cpu, memory uint16, podGroupID, nodeID string) error {

	return nil
}

// NewNode adds a new node to the node group
func (e *Expander) NewNode() (string, uint16, uint16, error) {
	return "", 0, 0, nil
}

func (e *Expander) DeletePod(ID string) error {
	return nil
}

// IsPodGhocked returns true if there is a ghost pod backing this real pod
// ghocked (ghost backed)
func (e *Expander) IsPodGhocked(ID string) bool {
	return true
}
