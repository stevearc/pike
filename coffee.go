package pike

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/stevearc/pike/plog"
)

// NewCoffee creates a Node that compiles coffeescript. Requires coffeescript
// (npm install -g coffee-script). There are up to three outputs:
//   1. js files
//   2. map files
//   3. coffee files
func NewCoffee() *Node {
	f := func(in, out []chan File) {
		useSourceMaps := len(out) > 1
		for file := range in[0] {
			if useSourceMaps {
				// We have to write the file to disk to get coffeescript to compile source maps
				basename := filepath.Base(file.Name())
				tempdir, err := ioutil.TempDir("", "")
				if err != nil {
					plog.Error("Error creating temporary directory")
					plog.Exc(err)
					continue
				}
				defer os.RemoveAll(tempdir)
				fullpath := filepath.Join(tempdir, basename)
				err = ioutil.WriteFile(fullpath, file.Data(), 0700)
				if err != nil {
					plog.Error("Error writing temporary file %q", fullpath)
					plog.Exc(err)
					continue
				}

				cmd := exec.Command("coffee", "-c", "-m", basename)
				cmd.Stderr = os.Stderr
				cmd.Dir = tempdir
				err = cmd.Run()
				if err != nil {
					plog.Error("error running coffee on %q", file.Name())
					continue
				}

				// Read in the javascript file
				jsfile := NewFile(file.Root(), file.Name(), nil)
				jsfile.SetExt(".js")
				jsFilePath := filepath.Join(tempdir, filepath.Base(jsfile.Name()))
				newData, err := ioutil.ReadFile(jsFilePath)
				if err != nil {
					plog.Error("Error reading file %q", jsFilePath)
					plog.Exc(err)
					continue
				}
				jsfile.SetData(newData)
				out[0] <- jsfile

				// Read in the mapfile
				mapfile := NewFile(file.Root(), file.Name(), nil)
				mapfile.SetExt(".map")
				mapFilePath := filepath.Join(tempdir, filepath.Base(mapfile.Name()))
				newData, err = ioutil.ReadFile(mapFilePath)
				if err != nil {
					plog.Error("Error reading file %q", mapFilePath)
					plog.Exc(err)
					continue
				}
				mapfile.SetData(newData)
				out[1] <- mapfile

				// Also send the original coffeescript file if needed
				if len(out) > 2 {
					out[2] <- file
				}
			} else {
				cmd := exec.Command("coffee", "-p", "-s")
				cmd.Stdin = bytes.NewReader(file.Data())
				cmd.Stderr = os.Stderr
				newData, err := cmd.Output()
				if err != nil {
					plog.Error("error running coffee on %q", file.Name())
					continue
				}
				file.SetData(newData)
				file.SetExt(".js")
				out[0] <- file
			}
		}
		for _, c := range out {
			close(c)
		}
	}
	runner := FxnRunnable(f)
	return NewNode("coffee", 1, 1, 1, 3, runner)
}
