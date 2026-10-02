package editgraph

import "testing"

func TestView(t *testing.T) {
	n := &Node{Name: "a"}
	g := Graph{Root: n}
	g.RootView().Name = "b"
	if g.Name() != "b" {
		t.Fatal(g.Name())
	}
	var _ interface{ Name() string } = Graph{}
}
