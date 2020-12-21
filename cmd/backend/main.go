package main

import (
	"fmt"
	j "github.com/AgentCoop/go-work"
	"github.com/AgentCoop/go-work-tcpbalancer/internal/backend"
	n "github.com/AgentCoop/go-work-tcpbalancer/internal/common/net"
	"github.com/jessevdk/go-flags"
	"os"
	"strconv"
)

func main() {
	backend.DefaultCliOptions()
	parser := flags.NewParser(&backend.CliOptions, flags.PassDoubleDash | flags.PrintErrors)
	parser.ParseArgs(os.Args)
	port := backend.CliOptions.Port

	if port == 0 {
		fmt.Printf("Specify a TCP port to listen on\n")
		os.Exit(-1)
	}

	localAddr := "127.0.0.1:" + strconv.Itoa(port)
	connManager := n.NewConnManager("tcp4", localAddr)

	mainJob := j.NewJob(connManager)
	mainJob.AddTask(n.ListenTask)

	if backend.CliOptions.Echo {
		mainJob.AddTask(backend.EchoService)
	}

	if backend.CliOptions.StressTest {
		mainJob.AddTask(backend.StressTestTask)
	}

	fmt.Printf("💻 server [ %s ] started on port %d\n", backend.CliOptions.Name, port)
	<-mainJob.Run()
}
