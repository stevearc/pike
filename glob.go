package pike

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/stevearc/pike/plog"
)

func remove(stringArr []string, str string) {
	for i, s := range stringArr {
		if s == str {
			stringArr[i] = ""
			break
		}
	}
}

// Glob creates a new source node that reads files from a directory. It
// will search recursively under 'root' for any files that match the patterns. The patterns are standard globs, with one exception. If you place a "!" at the beginning of the pattern, it will find all matching files and *remove* them from the existing set of matched files. You can use this, for example, to match all unminified css files:
//    n := pike.Glob("src", "*.css", "!*.min.css")
func Glob(root string, patterns ...string) *Node {
	sourceFunc := func(in, out []chan File) {
		paths := make([]string, 0, 10)
		for _, pattern := range patterns {
			if pattern[0] == '!' {
				for _, unmatch := range matchRecursive(root, pattern[1:]) {
					remove(paths, unmatch)
				}
			} else {
				paths = append(paths, matchRecursive(root, pattern)...)
			}
		}
		seenPaths := make(map[string]bool)
		for _, name := range paths {
			if name == "" || seenPaths[name] {
				continue
			}
			seenPaths[name] = true
			fullpath := filepath.Join(root, name)
			data, err := ioutil.ReadFile(fullpath)
			if err != nil {
				plog.Error("Error reading file %q", fullpath)
				continue
			}
			out[0] <- NewFile(root, name, data)
		}
		close(out[0])
	}
	runner := FxnRunnable(sourceFunc)
	return NewNode(fmt.Sprintf("%s -> %s", root, strings.Join(patterns, ":")), 0, 0, 1, 1, runner)
}

func matchRecursive(root, pattern string) []string {
	paths := make([]string, 0, 10)
	fullRoot := root
	subRoot, pattern := filepath.Split(pattern)
	if subRoot != "" {
		fullRoot = filepath.Join(root, subRoot)
	}
	filepath.Walk(fullRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			plog.Exc(err)
			return nil
		}
		// Ignore directories
		if info.IsDir() {
			return nil
		}
		subpath := path[len(root)+1:]
		matched, err := filepath.Match(pattern, filepath.Base(path))
		if err != nil {
			plog.Exc(err)
			return nil
		}
		if matched {
			paths = append(paths, subpath)
		}
		return nil
	})
	return paths
}
