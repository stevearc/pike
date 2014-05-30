package pike

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/stevearc/pike/plog"
)

// Write creates a node that writes files to a destination.
func Write(dest string) *Node {
	return WriteMode(dest, 0644)
}

// WriteMode creates a node that writes files to a destination with a
// specific file mode.
func WriteMode(dest string, perm os.FileMode) *Node {
	f := func(in, out chan File) {
		for file := range in {
			// Make sure the directory exists
			fullpath := filepath.Join(dest, file.Name())
			parent := filepath.Dir(fullpath)
			// Make sure directory is executable for any user with read perms
			// on the files contained within
			dirPerm := os.ModeDir | perm | 0100
			if perm&0040 > 0 {
				dirPerm |= 0010
			}
			if perm&0004 > 0 {
				dirPerm |= 0001
			}
			os.MkdirAll(parent, dirPerm)

			// Write the file
			plog.Info("Writing file %s", fullpath)
			err := ioutil.WriteFile(fullpath, file.Data(), perm)
			if err != nil {
				plog.Error("Error writing file %q", file.Name())
				plog.Exc(err)
			}

			// Pass the file on
			out <- file
		}
	}
	return NewFuncNode("write", f)
}
