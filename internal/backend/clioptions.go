package backend

var CliOptions struct {
	Port int `short:"p" description:"Listens on a port for inbound connections"`
	Echo bool `long:"echo" description:"Runs the server in an echo mode"`
}
