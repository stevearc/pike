package pike

import (
	"flag"
	"strings"
	"time"

	"github.com/stevearc/pike/plog"
)

func Start(graphMaker func(watch bool) []*Graph) {
	var watch bool
	var jsonFile string
	var prettyJson bool
	var interval int
	var level string

	flag.BoolVar(&watch, "w", false, "Rerun graphs constantly (should be used with ChangeFilters)")
	flag.StringVar(&jsonFile, "json", "", "The output file for json data (if using Json nodes)")
	flag.BoolVar(&prettyJson, "p", false, "Pretty-format the json data")
	flag.IntVar(&interval, "i", 200, "If using -w, sets the sleep interval between runs (in milliseconds)")
	flag.StringVar(&level, "l", "info", "Set the log level (debug, info, warn, error, fatal)")

	flag.Parse()

	if jsonFile != "" {
		SetJsonFile(jsonFile)
	}
	if prettyJson {
		SetJsonPretty(true)
	}

	switch strings.ToLower(level) {
	case "debug":
		plog.SetLevel(plog.DEBUG)
	case "info":
		plog.SetLevel(plog.INFO)
	case "warn":
		plog.SetLevel(plog.WARN)
	case "warning":
		plog.SetLevel(plog.WARN)
	case "error":
		plog.SetLevel(plog.ERROR)
	case "fatal":
		plog.SetLevel(plog.FATAL)
	default:
		plog.Fatal("Unrecognized log level %q", level)
	}

	graphs := graphMaker(watch)

	if watch {
		WatchAll(graphs, time.Duration(interval)*time.Millisecond)
	} else {
		RunAll(graphs)
	}
}
