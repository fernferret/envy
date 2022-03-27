package main

import (
	"encoding/json"
	"fmt"

	"github.com/fernferret/envy"
	"github.com/spf13/pflag"
)

type example struct {
	Url        string
	Once       bool
	Count      int
	CountFancy int
}

func main() {
	opts := example{}
	pflag.StringVar(&opts.Url, "url", "http://127.0.0.1:8080", "set the url")
	pflag.IntVar(&opts.Count, "count", 13, "a standard count")

	pflag.BoolVar(&opts.Once, "once", false, "perform the thing once and exit")
	pflag.IntVar(&opts.CountFancy, "count-fancy", 7, "a fancy count")

	envy.Disable("once")
	envy.Disable("count-fancy")

	envy.Parse("EXAMPLE")

	pflag.Parse()
	data, _ := json.MarshalIndent(opts, "", "  ")
	fmt.Println(string(data))
}
