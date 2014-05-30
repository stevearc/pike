package pike

import "github.com/stevearc/pike/plog"

// Debug creates a Node that prints the name of all files that pass through.
// It uses log level DEBUG, so if you don't see the messages make sure you do
// plog.SetLevel(plog.DEBUG).
func Debug(tag string) *Node {
	f := func(in, out chan File) {
		for file := range in {
			plog.Debug("%s: %s", tag, file.Name())
			out <- file
		}
	}
	return NewFuncNode("debug", f)
}
