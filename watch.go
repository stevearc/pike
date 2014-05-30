package pike

import "bytes"

// NewChangeFilter will only pass through files that have different data.
// Useful when you are running a Graph with Watch.
func NewChangeFilter() *Node {
	f := func(in, out []chan File, cache map[string]File) {
		for file := range in[0] {
			cachedFile, ok := cache[file.Name()]
			if ok && bytes.Equal(cachedFile.Data(), file.Data()) {
				continue
			}
			cache[file.Name()] = file.Copy()
			out[0] <- file
		}
		close(out[0])
	}
	runner := NewCacheRunnable(f)
	return NewNode("change filter", 1, 1, 1, 1, runner)
}

// Watches any number of streams. If any files change in either stream, it will
// pass on all files in the first stream. This is useful in the place of the
// ChangeFilter for files that implicitly depend on other files, such as a
// less file with @import.
func NewChangeWatcher() *Node {
	f := func(in, out []chan File, cache map[string]File) {
		primaryStream := make([]File, 0)
		anyChanges := false
		// Check primary stream for changes
		for file := range in[0] {
			cachedFile, ok := cache[file.Name()]
			if anyChanges || (ok && bytes.Equal(cachedFile.Data(), file.Data())) {
				continue
			}
			anyChanges = true
			cache[file.Name()] = file.Copy()
			primaryStream = append(primaryStream, file)
		}
		// Check all other input streams for changes
		for _, c := range in[1:] {
			for file := range c {
				cachedFile, ok := cache[file.Name()]
				if anyChanges || (ok && bytes.Equal(cachedFile.Data(), file.Data())) {
					continue
				}
				anyChanges = true
			}
		}
		if anyChanges {
			for _, f := range primaryStream {
				out[0] <- f
			}
		}
		close(out[0])
	}
	runner := NewCacheRunnable(f)
	return NewNode("change watcher", 2, -1, 1, 1, runner)
}

// NewChangeCache creates a Node that remembers all files that have passed
// through it, and replays them. Works well with ChangeFilter when you have
// later Nodes that must operate on all files.
func NewChangeCache() *Node {
	f := func(in, out []chan File, cache map[string]File) {
		seenFiles := make(map[string]bool)
		seenAny := false
		for file := range in[0] {
			seenFiles[file.Name()] = true
			cache[file.Name()] = file.Copy()
			out[0] <- file
			seenAny = true
		}
		if seenAny {
			for name, file := range cache {
				if !seenFiles[name] {
					out[0] <- file
				}
			}
		}
		close(out[0])
	}
	runner := NewCacheRunnable(f)
	return NewNode("change cache", 1, 1, 1, 1, runner)
}
