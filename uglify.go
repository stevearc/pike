package pike

import (
	"bytes"
	"os"
	"os/exec"

	"github.com/stevearc/pike/plog"
)

// Uglify creates a Node that runs uglifyjs on files. Requires uglifyjs (npm
// install -g uglify-js).
func Uglify() *Node {
	f := func(in, out chan File) {
		for file := range in {
			cmd := exec.Command("uglifyjs")
			cmd.Stdin = bytes.NewReader(file.Data())
			cmd.Stderr = os.Stderr
			newData, err := cmd.Output()
			if err != nil {
				plog.Error("error running uglifyjs on %q", file.Name())
				continue
			}
			file.SetData(newData)
			out <- file
		}
	}
	return NewFuncNode("uglify", f)
}
