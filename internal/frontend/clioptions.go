package frontend

var CliOptions struct {
	ProxyHost string `short:"h" required:"true" description:""`
	MaxConns int `long:"maxconns" description:""`
	MinNums int `long:"nmin" description:""`
	MaxNums int `long:"nmax" description:""`
}

func DefaultCliOptions() {
	CliOptions.MinNums = 1
	CliOptions.MaxNums = 10
	CliOptions.MaxConns = 1
}
