package sorter

// Item is the item to be sorted - representing pods in kubernetes, described with two dimensions cpu and memory
// plus set of additional parameters
type Item struct {
	ID      string
	CPU     uint16
	Memory  uint16
	Ghost   bool   //indicates if the pod is a ghost and can be kicked out any time and recreated later
	GroupID string //refers to either
}
