package backend

import (
	"github.com/jessevdk/go-flags"
	"os"
)

var CliOptions struct {
	CruncherPort int `long:"cruncher-port"`
	ImgResizePort int `long:"img-resize-port"`
	Name string `long:"name" required:"true" description:"Server name"`
}

func ParseCliOptions() {
	parser := flags.NewParser(&CliOptions, flags.PassDoubleDash | flags.PrintErrors)
	parser.ParseArgs(os.Args)
}
