package main

import (
	"fmt"
	"github.com/AgentCoop/go-work"
)

func initLogger() {
	job.DefaultLogLevel = CliOptions.LogLevel
	job.RegisterDefaultLogger(func() job.LogLevelMap {
		m := make(job.LogLevelMap, 2)
		handler := func(record interface{}, level int) {
			fmt.Printf(" 🌎 -> %s\n", record.(string))
		}
		m[0] = job.NewLogLevelMapItem(make(chan interface{}), handler)
		m[1] = job.NewLogLevelMapItem(make(chan interface{}), handler)
		return m
	})
}

