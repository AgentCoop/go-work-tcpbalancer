package frontend

var CliOptions struct {
	ProxyHost string `short:"h" required:"true" description:""`
	MaxConns int `long:"maxconns" description:""`
	MinBatchesPerConn int `long:"batch-min" description:""`
	MaxBatchesPerConn int `long:"batch-max" description:""`
	MinItemsPerBatch int `long:"batch-min-items" description:""`
	MaxItemsPerBatch int `long:"match-max-items" description:""`
}

func DefaultCliOptions() {
	CliOptions.MinBatchesPerConn = 1
	CliOptions.MaxBatchesPerConn = 1
	CliOptions.MinItemsPerBatch = 1
	CliOptions.MaxItemsPerBatch = 10
}
