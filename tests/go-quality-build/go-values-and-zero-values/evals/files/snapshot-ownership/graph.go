package editgraph

type Node struct {
	Name   string
	Values map[string][]byte
	Links  []*Node
}
type Graph struct{ Root *Node }

// RootView returns the original node for caller-owned in-place editing.
func (g Graph) RootView() *Node { return g.Root }
func (g Graph) Name() string {
	if g.Root == nil {
		return ""
	}
	return g.Root.Name
}
