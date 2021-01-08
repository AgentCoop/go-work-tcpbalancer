package main

import (
	"github.com/jessevdk/go-flags"
	"os"
)

var MainOptions struct {
	ProxyHost string `long:"proxy" required:"true"`
	LogLevel int `long:"loglevel"`
	MaxConns int `long:"maxconns"`
	ImgDir string `long:"input" required:"true"`
	OutputDir string `long:"output" required:"true"`
	Width uint `short:"w" required:"true"`
	Height uint `short:"h" required:"true"`
	Times int `long:"times"`
}

func newParser(data interface{}) *flags.Parser {
	parser := flags.NewParser(data, flags.PassDoubleDash | flags.PrintErrors | flags.IgnoreUnknown)
	return parser
}

func ParseCliOptions() {
	parser := newParser(&MainOptions)
	_, err := parser.ParseArgs(os.Args)
	if err != nil { panic(err) }

	if MainOptions.MaxConns == 0 { MainOptions.MaxConns = 1 }
	if MainOptions.Times == 0 { MainOptions.Times = 1 }
}
