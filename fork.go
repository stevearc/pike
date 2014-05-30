package pike

import "runtime"

// LoadBalancer creates a Node that allocates a portion of its files to each
// of its connected outputs. It will preserve the ordering of files, so the
// first output gets the first N files, the second output gets the second N
// files, etc.
func LoadBalancer() *Node {
	f := func(in, out []chan File) {
		files := make([]File, 0, 20)
		for file := range in[0] {
			files = append(files, file)
		}
		for i := 0; i < len(out); i++ {
			pieces := len(files) / len(out)
			i := i
			go func() {
				min := i * pieces
				max := (i + 1) * pieces
				if i == len(out)-1 {
					max = len(files)
				}
				for _, f := range files[min:max] {
					out[i] <- f
				}
				close(out[i])
			}()
		}
	}
	runner := FxnRunnable(f)
	return NewNode("load balancer", 1, 1, 1, -1, runner)
}

// LoadBalancerUnordered creates a Node that allocates a portion of
// its files to each of its connected outputs. There are no ordering
// guarantees.
func LoadBalancerUnordered() *Node {
	f := func(in, out []chan File) {
		idx := 0
		for file := range in[0] {
			out[idx] <- file
			idx = (idx + 1) % len(out)
		}
		for _, o := range out {
			close(o)
		}
	}
	runner := FxnRunnable(f)
	return NewNode("load balancer unordered", 1, 1, 1, -1, runner)
}

// Fork copies the target Node and uses a LoadBalancer to distribute
// the load among those nodes. It takes the output of each of these
// copies and merges them. If 'count' is 0, it will default to
// 2*runtime.NumCPU(). 'edges' is the number of output edges for the
// copied node. If you use 0, it will default to the max number of
// outputs of the node.
func (self *Node) Fork(nodeMaker Nodeable, count, edges int) *Node {
	return makeFork(self, nodeMaker.Node(), LoadBalancer(), Merge, count,
		edges)
}

// ForkUnordered is the same as Fork, but provides no ordering
// guarantees for the files. As such, it may provide a small speed
// boost.
func (self *Node) ForkUnordered(nodeMaker Nodeable, count, edges int) *Node {
	return makeFork(self, nodeMaker.Node(), LoadBalancerUnordered(),
		MergeUnordered, count, edges)
}

func makeFork(n1, n2, lb *Node, merge func() *Node, count, edges int) *Node {
	if count == 0 {
		count = 2 * runtime.NumCPU()
	}
	if edges == 0 {
		edges = n2.MaxOutputs
	}
	n1.Pipe(lb)
	fanIn := FanIn()
	merges := make([]*Node, edges, edges)
	for i := 0; i < edges; i++ {
		merges[i] = merge()
		if edges > 1 {
			merges[i].Pipe(fanIn)
		}
	}
	for i := 0; i < count; i++ {
		n2Copy := n2.Copy()
		p := lb.Pipe(n2Copy)
		for j := 0; j < edges; j++ {
			p.Pipe(merges[j])
		}
	}
	if edges == 1 {
		return merges[0]
	} else {
		return fanIn
	}
}
