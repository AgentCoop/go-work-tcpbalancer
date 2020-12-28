package frontend

import (
	"github.com/jessevdk/go-flags"
	"os"
)

type _MainOptions struct {
	ProxyHost string `long:"proxy" required:"true" description:""`
	Service string `long:"service"`
}

var MainOptions _MainOptions

var NumCruncherOptions struct {
	_MainOptions
	MaxConns int `long:"maxconns" description:""`
	MinBatchesPerConn int `long:"batch-min"`
	MaxBatchesPerConn int `long:"batch-max"`
	MinItemsPerBatch int `long:"batch-min-items"`
	MaxItemsPerBatch int `long:"batch-max-items"`
}

var ImgResizeOptions struct {
	_MainOptions
	ImgDir string `long:"input" required:"true"`
	OutputDir string `long:"output" required:"true"`
	Width uint `short:"w" required:"true"`
	Height uint `short:"h" required:"true"`
}

func ParseCliOptions() {
	parser := flags.NewParser(&MainOptions, flags.PassDoubleDash | flags.PrintErrors)
	parser.ParseArgs(os.Args)

	switch MainOptions.Service {
	case "cruncher":
		parser := flags.NewParser(&NumCruncherOptions, flags.PassDoubleDash | flags.PrintErrors)
		_, err := parser.ParseArgs(os.Args)
		if err != nil { panic(err) }
		// default values
		if NumCruncherOptions.MinBatchesPerConn == 0 {
			NumCruncherOptions.MinBatchesPerConn = 1
		}
		if NumCruncherOptions.MaxBatchesPerConn == 0 {
			NumCruncherOptions.MaxBatchesPerConn= NumCruncherOptions.MinBatchesPerConn
		}
		if NumCruncherOptions.MinItemsPerBatch == 0 {
			NumCruncherOptions.MinItemsPerBatch = 1
		}
		if NumCruncherOptions.MaxItemsPerBatch == 0 {
			NumCruncherOptions.MaxItemsPerBatch = NumCruncherOptions.MinItemsPerBatch + 10
		}
	case "imgresize":
		parser := flags.NewParser(&ImgResizeOptions, flags.PassDoubleDash | flags.PrintErrors)
		_, err := parser.ParseArgs(os.Args)
		if err != nil { panic(err) }
	default:
		os.Exit(-1)
	}
}
