package pike

import (
	"bytes"
	"path/filepath"
	"text/template"

	"github.com/stevearc/pike/plog"
)

// Rename creates a Node that renames the Files that pass through. 'format'
// is a go template string. The variables available in the template are below
// for the example filename "/app/src/myapp.js".
//   Fullname: /app/src/myapp.js
//   Dir     : /app/src
//   Name    : myapp.js
//   Barename: myapp
//   Ext     : .js
func Rename(format string) *Node {
	tmpl, err := template.New("rename").Parse(format)
	if err != nil {
		plog.Exc(err)
	}
	f := func(in, out chan File) {
		for file := range in {
			base := filepath.Base(file.Name())
			dir := filepath.Dir(file.Name())
			ext := filepath.Ext(base)
			bare := base[:len(base)-len(ext)]

			data := struct {
				Fullname string
				Dir      string
				Name     string
				Barename string
				Ext      string
			}{
				file.Name(),
				dir,
				base,
				bare,
				ext,
			}

			var buffer bytes.Buffer
			err := tmpl.Execute(&buffer, data)
			if err != nil {
				plog.Exc(err)
			} else {
				file.SetName(buffer.String())
				out <- file
			}
		}
	}
	return NewFuncNode("rename", f)
}
