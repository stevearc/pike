package pike

import "github.com/stevearc/pike/plog"

// NewFanin creates a node that takes inputs from multiple sources and maps
// them to its own outputs. This is used to converge multiple branches into a
// single node.
func NewFanIn() *Node {
	f := func(in, out []chan File) {
		if len(in) < len(out) {
			plog.Error("Fan-in node has fewer inputs than outputs")
			return
		}
		for i, c := range in {
			i := i
			c := c
			go func() {
				for file := range c {
					if i < len(out) {
						out[i] <- file
					}
				}
				if i < len(out) {
					close(out[i])
				}
			}()
		}
	}
	runner := FxnRunnable(f)
	return NewNode("fan-in", 1, -1, 1, -1, runner)
}

// Xargs copies the target node and connects each of the outputs of the base
// Node to one of the copies. If 'edges' is 0, Xargs will make an educated
// guess as to how many output edges to connect.
func (n1 *Node) Xargs(nodeMaker Nodeable, edges int) *Node {
	n2 := nodeMaker.Node()
	if edges == 0 {
		if n1.MaxOutputs == -1 {
			edges = len(n1.Inputs)
		} else {
			edges = n1.MaxOutputs
		}
	}
	if edges == 1 {
		return n1.Pipe(n2)
	}
	fanIn := NewFanIn()
	for i := 0; i < edges; i++ {
		n2Copy := n2.Copy()
		n1.Pipe(n2Copy)
		if n2.MaxOutputs != 0 {
			n2Copy.Pipe(fanIn)
		}
	}
	if n2.MaxOutputs != 0 {
		return fanIn
	} else {
		return nil
	}
}
