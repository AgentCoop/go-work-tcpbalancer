package main

import (
	"encoding/gob"
	"fmt"
	j "github.com/AgentCoop/go-work"
	n "github.com/AgentCoop/go-work-tcpbalancer/internal/common/net"
	"github.com/AgentCoop/go-work-tcpbalancer/internal/frontend"
	"github.com/jessevdk/go-flags"
	"os"
	"time"
)

func connToProxy(connManager n.ConnManager) {
	mainJob := j.NewJob(connManager)
	mainJob.WithPrerequisites(connManager.Connect(mainJob))
	mainJob.AddTask(frontend.SquareNumsInBatchTask)
	<-mainJob.Run()
	//go func(){
	//	select {
	//	case err := <- mainJob.GetError():
	//		fmt.Printf("err %s\n", err)
	//	}
	//}()
	time.Sleep(time.Second)
}

func main() {
	frontend.DefaultCliOptions()
	parser := flags.NewParser(&frontend.CliOptions, flags.PassDoubleDash | flags.PrintErrors)
	parser.ParseArgs(os.Args)

	fmt.Printf("Host: %s\n", frontend.CliOptions.ProxyHost)
	if len(frontend.CliOptions.ProxyHost) == 0 {
		fmt.Printf("Specify a proxy server to connect to\n")
		os.Exit(-1)
	}

	gob.Register(&frontend.CruncherPayload{})
	connManager := n.NewConnManager("tcp4", frontend.CliOptions.ProxyHost)
	for {
		//connToProxy(connManager)
		connToProxy(connManager)
		time.Sleep(time.Second)
	}
}
