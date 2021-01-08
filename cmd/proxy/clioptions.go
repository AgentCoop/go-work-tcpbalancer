package main

import (
	"github.com/jessevdk/go-flags"
	"os"
)

var CliOptions struct {
	Port int `short:"p" description:"Listens on a port for inbound connections"`
	UpstreamServers []string `short:"u" description:"Upstream servers"`
	LogLevel int `long:"loglevel"`
	MaxClientConns int `long:"maxconns"`
}

func newParser(data interface{}) *flags.Parser {
	parser := flags.NewParser(data, flags.PassDoubleDash | flags.PrintErrors | flags.IgnoreUnknown)
	return parser
}

func parseCliOptions() {
	parser := newParser(&CliOptions)
	_, err := parser.ParseArgs(os.Args)
	if err != nil { panic(err) }
}
