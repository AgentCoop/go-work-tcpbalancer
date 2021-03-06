package main

import (
	"github.com/jessevdk/go-flags"
	"os"
)

var CliOptions struct {
	Port int `long:"port" short:"p" required:"true"`
	UpstreamServers []string `short:"u" description:"Upstream servers"`
	LogLevel int `long:"loglevel"`
	MaxClientConns int `long:"maxconns"`
	Debug bool `long:"debug"`
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
