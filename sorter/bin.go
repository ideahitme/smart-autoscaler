package sorter

// Bin is a struct representing a container, i.e. a node described with two dimensional parameters cpu and memory
type Bin struct {
	ID     string
	CPU    uint16
	Memory uint16
	Items  []*Item //collection of pods in the node
}
