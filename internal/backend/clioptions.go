package backend

var CliOptions struct {
	Port int `short:"p" required:"true" description:"Listens on a port for inbound connections"`
	Echo bool `long:"echo" description:"Runs the server in an echo mode"`
	Name string `long:"name" required:"true" description:"Server name"`
}
