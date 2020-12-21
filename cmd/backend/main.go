package main

import (
	"fmt"
	j "github.com/AgentCoop/go-work"
	"github.com/AgentCoop/go-work-tcpbalancer/internal/backend"
	n "github.com/AgentCoop/go-work-tcpbalancer/internal/common/net"
	"github.com/jessevdk/go-flags"
	"os"
	"strconv"
	"strings"
)

func echoService(in []byte, c n.InboundConn) {
	msg := string(in)
	fmt.Printf("%s says %s\n", c.GetConn().RemoteAddr(), msg)
	upperCase := strings.ToUpper(msg)
	c.GetWriteChan() <- []byte(upperCase)
}

func main() {
	parser := flags.NewParser(&backend.CliOptions, flags.PassDoubleDash | flags.PrintErrors)
	parser.ParseArgs(os.Args)
	port := backend.CliOptions.Port

	if port == 0 {
		fmt.Printf("Specify a TCP port to listen on\n")
		os.Exit(-1)
	}

	localAddr := ":" + strconv.Itoa(port)
	connManager := n.NewConnManager("tcp4", localAddr)

	mainJob := j.NewJob(connManager)
	mainJob.AddTask(n.ListenTask)

	if backend.CliOptions.Echo {
		connManager.SetDataHandler(echoService)
	}

	<-mainJob.Run()
}
