package frontend

var CliOptions struct {
	ProxyHost string `short:"h" required:"true" description:""`
	ReqDataMaxLen int `long:"req-data-limit" description:"Maximum length of request data in bytes"`
	ReqMinTime int `long:"req-min-time" description:""`
	ReqMaxTime int `long:"req-max-time" description:""`
	MaxConns int `long:"maxconns" description:""`
}

func DefaultCliOptions() {
	CliOptions.ReqDataMaxLen = 1024
	CliOptions.ReqMinTime = 50
	CliOptions.ReqMaxTime = 1000
	CliOptions.MaxConns = 1
}
