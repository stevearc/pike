package pike

import (
	"fmt"
	"strings"
)

// NewHtmlTemplateCache compiles html files into an angular.run() command that
// puts the html into the $templateCache.
func NewHtmlTemplateCache(module, prefix string) *Node {
	prefix = "/" + strings.Trim(prefix, "/") + "/"
	if prefix == "//" {
		prefix = "/"
	}
	f := func(in, out chan File) {
		for file := range in {
			file.SetData([]byte(fmt.Sprintf(`angular.module('%s').run(['$templateCache', function($templateCache) {
	$templateCache.put('%s%s', %q);
}]);`, module, prefix, file.Name(), file.Data())))
			file.SetExt("_tmpl.js")
			out <- file
		}
	}
	return NewFunc("html2tc", f)
}
