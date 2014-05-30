// Pike is an asset pipeline and make tool
package pike

import (
	"fmt"
	"strings"
)

// Node is a component of the asset graphs. Each node performs an operation on
// the files as they pass through, and can be connected to other Nodes.
type Node struct {
	Name       string
	Inputs     []*Node
	Outputs    []*Node
	MinInputs  int
	MaxInputs  int
	MinOutputs int
	MaxOutputs int
	Runner     Runnable
}

// Nodeable is an interface that can be converted to a Node. This is useful for
// connection methods such as Pipe, which can be applied to a Node or a Graph
// (both of which implement Nodeable).
type Nodeable interface {
	Node() *Node
}

func nodeIndex(nodes []*Node, node *Node) int {
	for i, v := range nodes {
		if v == node {
			return i
		}
	}
	return -1
}

func (node *Node) String() string {
	return fmt.Sprintf("Node(%q)", node.Name)
}

// NewNode constructs a Node struct.
func NewNode(name string, minIn, maxIn, minOut, maxOut int, runner Runnable) *Node {
	return &Node{name, nil, nil, minIn, maxIn, minOut, maxOut, runner}
}

// NewFuncNode constructs a simple 1-input, 1-output node from a function.
func NewFuncNode(name string, run func(in, out chan File)) *Node {
	f := func(in, out []chan File) {
		run(in[0], out[0])
		close(out[0])
	}
	runner := FxnRunnable(f)
	return NewNode(name, 1, 1, 1, 1, runner)
}

// Create a deep copy of a Node. Note that this will reset the Inputs and
// Outputs.
func (node *Node) Copy() *Node {
	return NewNode(node.Name, node.MinInputs, node.MaxInputs, node.MinOutputs,
		node.MaxOutputs, node.Runner.Copy())
}

// For Node this is a no-op
func (self *Node) Node() *Node {
	return self
}

// Walk this node and all parents.
func (n *Node) WalkUp(out chan *Node) {
	n._walkUp(out)
	close(out)
}
func (n *Node) _walkUp(out chan *Node) {
	out <- n
	for _, parent := range n.Inputs {
		parent._walkUp(out)
	}
}

// Walk this node and all children.
func (n *Node) Walk(out chan *Node) {
	n._walk(out)
	close(out)
}
func (n *Node) _walk(out chan *Node) {
	out <- n
	for _, child := range n.Outputs {
		child._walk(out)
	}
}

// Pipe creates a connection from one Node to a Node or Graph.
func (n1 *Node) Pipe(nodeMaker Nodeable) *Node {
	n2 := nodeMaker.Node()
	n1.Outputs = append(n1.Outputs, n2)
	n2.Inputs = append(n2.Inputs, n1)
	return n2
}

// Dot returns the dot representation of this node and all outbound edges.
func (self *Node) Dot(indent string) string {
	lines := make([]string, 0, 2)

	lines = append(lines, fmt.Sprintf("%s\"%p\" [label=%q];", indent, self, self.Name))
	for i, next := range self.Outputs {
		dotEdge := fmt.Sprintf("%s\"%p\" -> \"%p\"", indent, self, next)
		if len(self.Outputs) > 1 {
			dotEdge += fmt.Sprintf(" [label=\"%d\"]", i)
		}
		lines = append(lines, dotEdge)
	}

	return strings.Join(lines, fmt.Sprintf("\n%s", indent))
}
