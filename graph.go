package pike

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/stevearc/pike/plog"
)

// Graph is a collection of Nodes and edges between them.
type Graph struct {
	Name   string
	nodes  []*Node
	Source *Node
	Sink   *Node
}

// NewGraph is a simple constructor for Graph
func NewGraph(name string) *Graph {
	return &Graph{name, make([]*Node, 0, 10), nil, nil}
}

// Add will add any number of nodes to a Graph. The Graph will also
// walk the nodes' ancestors and children and add those Nodes as well.
func (graph *Graph) Add(nodes ...*Node) {
	allNodes := make(map[*Node]bool)
	for _, n := range graph.nodes {
		allNodes[n] = true
	}
	for _, n := range nodes {
		children := make(chan *Node)
		go n.Walk(children)
		for child := range children {
			allNodes[child] = true
		}
		parents := make(chan *Node)
		go n.WalkUp(parents)
		for parent := range parents {
			allNodes[parent] = true
		}
	}

	for n := range allNodes {
		graph.nodes = append(graph.nodes, n)
	}
}

// Check the input and output edges for constraint violations.
func (graph *Graph) validate() error {
	for _, n := range graph.nodes {
		if n.MaxInputs >= 0 && len(n.Inputs) > n.MaxInputs {
			return errors.New(fmt.Sprintf("%v has too many inputs", n))
		}
		if len(n.Inputs) < n.MinInputs && n != graph.Source {
			return errors.New(fmt.Sprintf("%v has too few inputs", n))
		}
		if n.MaxOutputs >= 0 && len(n.Outputs) > n.MaxOutputs {
			return errors.New(fmt.Sprintf("%v has too many outputs", n))
		}
		if len(n.Outputs) != 0 && len(n.Outputs) < n.MinOutputs && n != graph.Sink {
			return errors.New(fmt.Sprintf("%v has too few outputs", n))
		}
	}
	return nil
}

// Run will start running a Graph. It will return a WaitGroup that can
// be used to detect when the Graph has processed all files.
func (graph *Graph) Run() (*sync.WaitGroup, error) {
	if err := graph.validate(); err != nil {
		return nil, err
	}
	if graph.Source != nil {
		return nil, errors.New("Cannot run a graph with a source!")
	}
	if graph.Sink != nil {
		return nil, errors.New("Cannot run a graph with a sink!")
	}
	return graph.start(make([]chan File, 0), make([]chan File, 0))
}

func (graph *Graph) start(in, out []chan File) (*sync.WaitGroup, error) {
	if err := graph.validate(); err != nil {
		return nil, err
	}
	inMap := make(map[*Node][]chan File)
	outMap := make(map[*Node][]chan File)

	waitGroup := &sync.WaitGroup{}
	// First pass creates the channel slices
	for _, n := range graph.nodes {
		inMap[n] = make([]chan File, len(n.Inputs))
		// If node has dangling outputs, redirect them to channels that
		// consume and discard
		if n != graph.Sink && len(n.Outputs) == 0 && n.MinOutputs > 0 {
			numOutputs := n.MaxOutputs
			if numOutputs == -1 {
				numOutputs = len(n.Inputs)
			}
			outMap[n] = make([]chan File, numOutputs)
			for i := 0; i < numOutputs; i++ {
				c := make(chan File, 10)
				outMap[n][i] = c
				waitGroup.Add(1)
				go func() {
					for _ = range c {
					}
					waitGroup.Done()
				}()
			}
		} else {
			outMap[n] = make([]chan File, len(n.Outputs))
		}
	}

	// Second pass creates the channels
	for _, n := range graph.nodes {
		for i, input := range n.Inputs {
			j := nodeIndex(input.Outputs, n)
			c := make(chan File, 10)
			outMap[input][j] = c
			inMap[n][i] = c
		}
	}

	// Set the input channels for the source
	if graph.Source != nil {
		inMap[graph.Source] = in
	}
	// Set the output channels for the sink
	if graph.Sink != nil {
		outMap[graph.Sink] = out
	}

	// Start a goroutine for each Node
	for _, n := range graph.nodes {
		n := n
		waitGroup.Add(1)
		go func() {
			n.Runner.Run(inMap[n], outMap[n])
			waitGroup.Done()
		}()
	}
	return waitGroup, nil
}

// Watch will run the Graph continuously, sleeping for 'poll' between
// runs. If you use this method, your Graph should contain some nodes
// that watch for file changes (i.e. NewChangeFilter), otherwise it
// will just continually process all your files.
func (graph *Graph) Watch(poll time.Duration, quit chan int) error {
	for {
		waitGroup, err := graph.Run()
		if err != nil {
			return err
		}
		waitGroup.Wait()

		select {
		case <-quit:
			return nil
		default:
			time.Sleep(poll)
		}
	}
}

// Copy creates a deep copy of the Graph.
func (self *Graph) Copy() *Graph {
	newNodes := make([]*Node, len(self.nodes))
	var source, sink *Node

	nodeMap := make(map[*Node]*Node)
	for i, node := range self.nodes {
		newNode := node.Copy()
		nodeMap[node] = newNode
		newNodes[i] = newNode
		if node == self.Source {
			source = newNode
		} else if node == self.Sink {
			sink = newNode
		}
	}
	for _, node := range self.nodes {
		newNode := nodeMap[node]
		for _, in := range node.Inputs {
			newNode.Inputs = append(newNode.Inputs, nodeMap[in])
		}
		for _, out := range node.Outputs {
			newNode.Outputs = append(newNode.Outputs, nodeMap[out])
		}
	}

	return &Graph{self.Name, newNodes, source, sink}
}

// GraphRunnable is a Runnable that delegates to a Graph.
type GraphRunnable struct {
	Fxn   func(in, out []chan File, graph *Graph)
	Graph *Graph
}

func (self *GraphRunnable) Run(in, out []chan File) {
	self.Fxn(in, out, self.Graph)
}

func (self *GraphRunnable) Copy() Runnable {
	return &GraphRunnable{self.Fxn, self.Graph.Copy()}
}

// Creates a Node that wraps the Graph. This allows you to use Graphs
// as a single Node inside other Graphs.
func (self *Graph) Node() *Node {
	f := func(in, out []chan File, graph *Graph) {
		waitGroup, err := graph.start(in, out)
		if err != nil {
			plog.Exc(err)
			return
		}
		waitGroup.Wait()
	}
	minIn, maxIn, minOut, maxOut := 0, 0, 0, 0
	if self.Source != nil {
		minIn = self.Source.MinInputs
		maxIn = self.Source.MaxInputs
	}
	if self.Sink != nil {
		minOut = self.Sink.MinOutputs
		maxOut = self.Sink.MaxOutputs
	}
	runner := &GraphRunnable{f, self.Copy()}
	return NewNode(fmt.Sprintf("graph(%q)", self.Name), minIn, maxIn,
		minOut, maxOut, runner)
}

// Dot returns the dot representation of the Graph. If indent is not
// "", it will render this graph in the format of a subgraph.
func (self *Graph) Dot(indent string) string {
	re := regexp.MustCompile("[^A-Za-z0-9_\\-]")
	name := re.ReplaceAll([]byte(self.Name), []byte("_"))

	lines := make([]string, 0, 20)

	if len(indent) > 0 {
		lines = append(lines, fmt.Sprintf("%ssubgraph cluster_%s {",
			indent, name))
		lines = append(lines, fmt.Sprintf("%s  label = \"%s\";",
			indent, self.Name))
	} else {
		lines = append(lines, fmt.Sprintf("%sdigraph %s {",
			indent, name))
	}
	for _, node := range self.nodes {
		lines = append(lines, node.Dot(indent+"  "))
	}
	lines = append(lines, "}")
	return strings.Join(lines, fmt.Sprintf("\n%s", indent))
}

// Render will use the dot format to render an image for the Graph. The
// file type is determined by the extension on 'outfile'. Requires
// graphviz.
func (self *Graph) Render(outfile string) error {
	ext := filepath.Ext(outfile)
	imageFormat := ext[1:]

	dotFile, err := ioutil.TempFile("", "graph")
	if err != nil {
		return err
	}
	defer dotFile.Close()
	defer os.Remove(dotFile.Name())
	_, err = dotFile.Write([]byte(self.Dot("")))
	if err != nil {
		return err
	}

	cmd := exec.Command("dot", "-T", imageFormat, "-o", outfile, dotFile.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		plog.Error("Dot failed. Is graphviz installed?")
		return err
	}
	return nil
}

// RunAll runs a slice of Graphs and blocks until they all complete.
func RunAll(graphs []*Graph) {
	groups := make([]*sync.WaitGroup, 0, 10)
	for _, g := range graphs {
		waitGroup, err := g.Run()
		if err != nil {
			plog.Exc(err)
		} else {
			groups = append(groups, waitGroup)
		}
	}
	for _, wg := range groups {
		wg.Wait()
	}
}

// WatchAll will run a slice of Graphs continuously until the program
// quits.
func WatchAll(graphs []*Graph, poll time.Duration) {
	for {
		RunAll(graphs)
		time.Sleep(poll)
	}
}
