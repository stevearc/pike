package pike

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/stevearc/pike/plog"
)

// NewLess creates a Node that runs the LESS CSS preprocessor on files.
// Requires less (npm install -g less)
func NewLess() *Node {
	f := func(in, out chan File) {
		for file := range in {
			cmd := exec.Command("lessc", "-")
			cmd.Stdin = bytes.NewReader(file.Data())
			cmd.Stderr = os.Stderr
			cmd.Dir = filepath.Dir(file.Fullpath())
			newData, err := cmd.Output()
			if err != nil {
				plog.Error("error running lessc on %q", file.Name())
				plog.Exc(err)
				continue
			}
			file.SetData(newData)
			file.SetExt(".css")
			out <- file
		}
	}
	return NewFunc("less", f)
}
