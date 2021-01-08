package main

import (
	"github.com/jessevdk/go-flags"
	"os"
)

var MainOptions struct {
	ProxyHost string `long:"proxy" required:"true"`
	Service string `long:"service" required:"true"`
	LogLevel int `long:"loglevel"`
	MaxConns int `long:"maxconns"`
}

var CruncherOpts struct {
	MinBatchesPerConn uint `long:"batch-min"`
	MaxBatchesPerConn uint `long:"batch-max"`
	MinItemsPerBatch uint `long:"batch-min-items"`
	MaxItemsPerBatch uint `long:"batch-max-items"`
}

var ImgResizeOpts struct {
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
	remOpts, err := parser.ParseArgs(os.Args)
	if err != nil { panic(err) }

	if MainOptions.MaxConns == 0 { MainOptions.MaxConns = 1 }

	switch MainOptions.Service {
	case "cruncher":
		parser := newParser(&CruncherOpts)
		_, err := parser.ParseArgs(remOpts)
		if err != nil { panic(err) }
		// default values
		if CruncherOpts.MinBatchesPerConn == 0 {
			CruncherOpts.MinBatchesPerConn = 1
		}
		if CruncherOpts.MaxBatchesPerConn == 0 {
			CruncherOpts.MaxBatchesPerConn= CruncherOpts.MinBatchesPerConn
		}
		if CruncherOpts.MinItemsPerBatch == 0 {
			CruncherOpts.MinItemsPerBatch = 1
		}
		if CruncherOpts.MaxItemsPerBatch == 0 {
			CruncherOpts.MaxItemsPerBatch = CruncherOpts.MinItemsPerBatch + 10
		}
	case "imgresize":
		parser := newParser(&ImgResizeOpts)
		_, err := parser.ParseArgs(remOpts)
		if err != nil { panic(err) }
		if ImgResizeOpts.Times == 0 { ImgResizeOpts.Times = 1 }
	default:
		panic("invalid service name")
	}
}
