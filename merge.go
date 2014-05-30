package pike

import "time"

// NewMerge creates a Node that merges multiple input streams into a single
// output. The ordering of the edges is preserved (i.e. all files from the
// first edge will preceed files from the second input edge).
func NewMerge() *Node {
	f := func(in, out []chan File) {
		o := out[0]
		for _, c := range in {
			for file := range c {
				o <- file
			}
		}
		close(out[0])
	}
	runner := FxnRunnable(f)
	return NewNode("merge", 2, -1, 1, 1, runner)
}

// NewMergeUnordered creates a Node that merges multiple input streams into a
// single output. There are no ordering guarantees across edges.
func NewMergeUnordered() *Node {
	f := func(in, out []chan File) {
		allClosed := false
		for !allClosed {
			allClosed = true
			for _, c := range in {
				select {
				case f, ok := <-c:
					if ok {
						out[0] <- f
						allClosed = false
					}
				default:
					allClosed = false
				}
			}
			time.Sleep(10 * time.Millisecond)
		}
		close(out[0])
	}
	runner := FxnRunnable(f)
	return NewNode("merge", 2, -1, 1, 1, runner)
}
