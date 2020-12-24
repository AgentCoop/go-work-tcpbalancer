package frontend

import (
	"github.com/jessevdk/go-flags"
	"os"
)

var CliOptions struct {
	ProxyHost string `short:"h" required:"true" description:""`
	MaxConns int `long:"maxconns" description:""`
	MinBatchesPerConn int `long:"batch-min" description:""`
	MaxBatchesPerConn int `long:"batch-max" description:""`
	MinItemsPerBatch int `long:"batch-min-items" description:""`
	MaxItemsPerBatch int `long:"batch-max-items" description:""`
}

func ParseCliOptions() {
	parser := flags.NewParser(&CliOptions, flags.PassDoubleDash | flags.PrintErrors)
	parser.ParseArgs(os.Args)

	if CliOptions.MinBatchesPerConn == 0 {
		CliOptions.MinBatchesPerConn = 1
	}
	if CliOptions.MaxBatchesPerConn == 0 {
		CliOptions.MaxBatchesPerConn= CliOptions.MinBatchesPerConn
	}

	if CliOptions.MinItemsPerBatch == 0 {
		CliOptions.MinItemsPerBatch = 1
	}
	if CliOptions.MaxItemsPerBatch == 0 {
		CliOptions.MaxItemsPerBatch = CliOptions.MinItemsPerBatch + 10
	}
}
