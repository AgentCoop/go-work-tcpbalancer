package backend

import (
	"github.com/jessevdk/go-flags"
	"os"
)

var CliOptions struct {
	Port int `long:"port" short:"p"`
	Service string `long:"service"`
	Name string `long:"name" required:"true" description:"Server name"`
	CpuProfile string `long:"cpuprofile"`
}

func ParseCliOptions() {
	parser := flags.NewParser(&CliOptions, flags.PassDoubleDash | flags.PrintErrors)
	parser.ParseArgs(os.Args)
}
