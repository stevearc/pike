package pike

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"path/filepath"
)

// Fingerprint creates a Node that will add an md5 hash to the name of all
// files it processes. This is useful for cache busting.
func Fingerprint() *Node {
	f := func(in, out chan File) {
		for file := range in {
			sum := md5.Sum(file.Data())
			hash := hex.EncodeToString(sum[:])
			basename := filepath.Base(file.Name())
			parent := filepath.Dir(file.Name())
			ext := filepath.Ext(file.Name())
			bareName := basename[:len(basename)-len(ext)]

			newname := fmt.Sprintf("%s-%s%s", bareName, hash, ext)
			file.SetName(filepath.Join(parent, newname))
			out <- file
		}
	}
	return NewFuncNode("fingerprint", f)
}
