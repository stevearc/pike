package pike

// Concat creates a node that will concatenate all processed files into a
// single file.
func Concat(path string) *Node {
	f := func(in, out chan File) {
		bigFile := NewFile("", path, make([]byte, 0))
		for file := range in {
			bigFile.SetData(append(bigFile.Data(), file.Data()...))
			bigFile.SetData(append(bigFile.Data(), []byte("\n")...))
		}
		if len(bigFile.Data()) > 0 {
			out <- bigFile
		}
	}
	return NewFuncNode("concat", f)
}
