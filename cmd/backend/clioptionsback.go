package main

import (
	"github.com/jessevdk/go-flags"
	"os"
)

var CliOptions struct {
	Port int `long:"port" short:"p" required:"true"`
	Name string `long:"name" required:"true" description:"Server name"`
	CpuProfile string `long:"cpuprofile"`
	Debug bool `long:"debug"`
	LogLevel int `long:"loglevel"`
}

func parseCliOptions() {
	parser := flags.NewParser(&CliOptions, flags.PassDoubleDash | flags.PrintErrors | flags.IgnoreUnknown)
	_, err := parser.ParseArgs(os.Args)
	if err != nil { panic(err) }
}
