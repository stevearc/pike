package pike

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/stevearc/pike/plog"
)

var cache = struct {
	AllFiles map[string]map[string]bool
	Output   map[string][]string
}{
	make(map[string]map[string]bool),
	make(map[string][]string),
}

var config = struct {
	Lock        *sync.Mutex
	Destination string
	Pretty      bool
}{
	&sync.Mutex{},
	"",
	false,
}

// SetJsonFile sets the global json output file. This must be set in order to
// use Json nodes.
func SetJsonFile(dest string) {
	config.Destination = dest
}

// SetJsonPretty will optionally dump the json in a human-readable, indented
// format.
func SetJsonPretty(indent bool) {
	config.Pretty = indent
}

func writeJsonFile() {
	config.Lock.Lock()
	if config.Destination == "" {
		plog.Error("Json file not set. Use node.SetJsonFile()")
	} else {
		// Make sure the directory exists
		parent := filepath.Dir(config.Destination)
		os.MkdirAll(parent, os.ModeDir|0755)

		// Write the file
		plog.Info("Dumping json %q", config.Destination)
		var err error
		var jsonData []byte
		if config.Pretty {
			jsonData, err = json.MarshalIndent(cache.Output, "", "  ")
		} else {
			jsonData, err = json.Marshal(cache.Output)
		}
		if err != nil {
			plog.Exc(err)
		}
		err = ioutil.WriteFile(config.Destination, jsonData, 0644)
		if err != nil {
			plog.Error("Error writing file %q", config.Destination)
			plog.Exc(err)
		}
	}
	config.Lock.Unlock()
}

// NewJson creates a Node that dumps the paths of all files into a json file.
// The Json file is global (it's the same for ALL graphs in a process) and must
// be set with SetJsonFile.
func NewJson(key string) *Node {
	f := func(in, out chan File) {
		newFiles := false
		for file := range in {
			newFiles = true
			_, ok := cache.AllFiles[key]
			if !ok {
				cache.AllFiles[key] = make(map[string]bool)
				cache.Output[key] = make([]string, 0, 10)
			}

			// Add file to output if it's not already in output
			_, ok = cache.AllFiles[key][file.Name()]
			if !ok {
				cache.AllFiles[key][file.Name()] = true
				cache.Output[key] = append(cache.Output[key], file.Name())
			}
			out <- file
		}
		if newFiles {
			writeJsonFile()
		}
	}
	return NewFunc("json", f)
}
