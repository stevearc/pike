Pike
====
An asset pipeline and make tool.

## Getting Started

First install Go and configure your GOPATH (see http://golang.org/doc/install).
Then install pike with:

```
    go get github.com/stevearc/pike
```

Pike represents all operations as [directed acyclic
graphs](http://en.wikipedia.org/wiki/Directed_acyclic_graph). These graphs are
comprised of nodes, each of which performs a single, isolated operation. Most
use cases will only require simple, linear graphs. Let's look at an example
that compiles and minifies LESS files:

```
n := pike.Glob("app/src", "*.less")
n = n.Pipe(pike.Less())
n = n.Pipe(pike.CleanCss())
n = n.Pipe(pike.Write("build"))

g := pike.NewGraph("app.less")
g.Add(n)
```

Now let's look at a more complicated example. Coffeescript compiles into
javascript, but it can also produce source maps for improved debugging. The
coffeescript node has multiple outputs for javascript, source maps, and the
original coffeescript files. Let's look at an example that takes advantage of
that.

```
n := pike.Glob("app/src", "*.coffee")

// We use Fork instead of Pipe to take advantage of multiple cores
// The '0' will cause it to create 2*NumCPUs nodes
// The '3' is the number of output edges on the terminus
n = n.Fork(pike.Coffee(), 0, 3)

// Since we have 3 output edges, we use Xargs here instead of Pipe
n = n.Xargs(pike.Write("build"), 3)

g := pike.NewGraph("app.js")
g.Add(n)
```

## A Full Example

Let's combine both of the previous graphs into a real file you might use in
production.

```
package main

func makeAllGraphs(watch bool) []*pike.Graph {
	allGraphs := make([]*pike.Graph, 0, 10)

	n := pike.Glob("app/src", "*.less")
	// If we're watching for changes, make sure the graph only sends through
	// files that have changed.
	if watch {
		n = n.Pipe(pike.ChangeFilter())
	}
	n = n.Pipe(pike.Less())
	n = n.Pipe(pike.CleanCss())
	// Write all output file names to the json file
	n = n.Pipe(pike.Json("app.css"))
	n = n.Pipe(pike.Write("build"))
	g := pike.NewGraph("app.less")
	g.Add(n)
	allGraphs = append(allGraphs, g)

	n = pike.Glob("app/src", "*.coffee")
	if watch {
		n = n.Pipe(pike.ChangeFilter())
		n = n.Fork(pike.Coffee(), 0, 3)
	} else {
		// If we're not watching for changes, discard source maps
		n = n.Fork(pike.Coffee(), 0, 1)
		// Merge all js files into a single file
		n = n.Pipe(pike.Concat("app.js"))
		// Minify the final file
		n = n.Pipe(pike.Uglify())
	}
	n = n.Xargs(pike.Write("build"), 0)
	n = n.Pipe(pike.Json("app.js"))
	g = pike.NewGraph("app.js")
	g.Add(n)
	allGraphs = append(allGraphs, g)

	return allGraphs
}

func main() {
	pike.Start(makeAllGraphs)
}
```

You can run this file with `go run build.go`. It accepts commandline arguments.
For more details run `go run build.go -h`.

## Integration

After you build your assets, you will likely need to integrate them somehow
with your application. Pike allows you to dump a list of all generated files
into a json file, which can then be read by your application. Just call
`pike.SetJsonFile("out.json")` and put a `pike.Json("app.js")` in each
graph.

## Debugging

If you run into problems with your graphs, it may be useful to visualize what
the graphs are doing. You can use the Graph.Dot() method to print it out as dot
syntax, or call Graph.Render("mygraph.png") to see an image of what your graph
looks like. At the moment it does not display subgraphs very well. Sorry.
