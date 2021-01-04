package main

import (
	"github.com/jessevdk/go-flags"
	"os"
)

var CliOptions struct {
	Port int `long:"port" short:"p" required:"true"`
	Service string `long:"service"`
	Name string `long:"name" required:"true" description:"Server name"`
	CpuProfile string `long:"cpuprofile"`
	Debug bool `long:"debug"`
	LogLevel int `long:"loglevel"`
}

func ParseCliOptions() {
	parser := flags.NewParser(&CliOptions, flags.PassDoubleDash | flags.PrintErrors)
	_, err := parser.ParseArgs(os.Args)
	if err != nil { panic(err) }
}
