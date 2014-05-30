package pike

import "sync"

// Sink creates a Node that terminates a branch. It consumes and discards
// all files it receives.
func Sink() *Node {
	f := func(in, out []chan File) {
		waitGroup := &sync.WaitGroup{}
		for _, c := range in {
			c := c
			waitGroup.Add(1)
			go func() {
				for _ = range c {
				}
				waitGroup.Done()
			}()
		}
		waitGroup.Wait()
	}
	runner := FxnRunnable(f)
	return NewNode("sink", 1, -1, 0, 0, runner)
}
