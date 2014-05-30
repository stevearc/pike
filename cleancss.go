package pike

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/stevearc/pike/plog"
)

// NewCleanCss creates a node that runs cleancss on files. Requires cleancss
// (npm install -g clean-css).
func NewCleanCss() *Node {
	f := func(in, out chan File) {
		for file := range in {
			path := filepath.Dir(file.Fullpath())
			cmd := exec.Command("cleancss", "--root", path)
			cmd.Stdin = bytes.NewReader(file.Data())
			cmd.Stderr = os.Stderr
			cmd.Dir = path
			newData, err := cmd.Output()
			if err != nil {
				plog.Error("error running cleancss on %q", file.Name)
				plog.Exc(err)
				continue
			}
			file.SetData(newData)
			out <- file
		}
	}
	return NewFunc("cleancss", f)
}
