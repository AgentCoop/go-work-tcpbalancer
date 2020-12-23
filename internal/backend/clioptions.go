package backend

var CliOptions struct {
	Port int `short:"p" required:"true" description:"Listens on a port for inbound connections"`
	Echo bool `long:"echo" description:"Runs the server in an echo mode"`
	Name string `long:"name" required:"true" description:"Server name"`
	RespDataMaxLen int `long:"res-data-limit" description:"Maximum length of response data in bytes"`
	RespMinTime int `long:"res-min-time" description:""`
	RespMaxTime int `long:"res-max-time" description:""`
}

func DefaultCliOptions() {
	CliOptions.RespDataMaxLen = 32
	CliOptions.RespMinTime = 100
	CliOptions.RespMaxTime = 2000
}
