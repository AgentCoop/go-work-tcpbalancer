package main

import (
	"github.com/jessevdk/go-flags"
	"os"
)

var cliOptions struct {
	ProxyHost string `long:"proxy" required:"true"`
	LogLevel int `long:"loglevel"`
	MaxConns int `long:"maxconns"`
	MinConns int `long:"minconns"`
	ImgDir string `long:"input" required:"true"`
	OutputDir string `long:"output" required:"true"`
	Width uint `short:"w" required:"true"`
	Height uint `short:"h" required:"true"`
	Times int `long:"times"`
	DryRun bool `long:"dry-run"`
	Debug bool `long:"debug"`
}

func newParser(data interface{}) *flags.Parser {
	parser := flags.NewParser(data, flags.PassDoubleDash | flags.PrintErrors | flags.IgnoreUnknown)
	return parser
}

func ParseCliOptions() {
	parser := newParser(&cliOptions)
	_, err := parser.ParseArgs(os.Args)
	if err != nil { panic(err) }

	if cliOptions.MaxConns == 0 { cliOptions.MaxConns = 1 }
	if cliOptions.MinConns == 0 { cliOptions.MinConns = 1 }
	if cliOptions.Times == 0 { cliOptions.Times = 1 }
}
