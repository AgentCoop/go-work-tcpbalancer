package main

import (
	"fmt"
	j "github.com/AgentCoop/go-work"
	n "github.com/AgentCoop/go-work-tcpbalancer/internal/common/net"
	"github.com/AgentCoop/go-work-tcpbalancer/internal/frontend"
	"github.com/jessevdk/go-flags"
	"os"
	"time"
)

func main() {
	frontend.DefaultCliOptions()
	parser := flags.NewParser(&frontend.CliOptions, flags.PassDoubleDash | flags.PrintErrors)
	parser.ParseArgs(os.Args)
	if len(frontend.CliOptions.ProxyHost) == 0 {
		fmt.Printf("Specify a proxy server to connect to\n")
		os.Exit(-1)
	}

	connManager := n.NewConnManager("tcp4", frontend.CliOptions.ProxyHost)
	for {
		mainJob := j.NewJob(connManager)
		fmt.Printf("#1\n")
		mainJob.WithPrerequisites(connManager.Connect(mainJob))
		fmt.Printf("#2\n")
		mainJob.AddTask(frontend.ClientTask)
		fmt.Printf("#3\n")
		<-mainJob.Run()
		go func(){
			select {
			case err := <- mainJob.GetError():
				fmt.Printf("err %s\n", err)
			}
		}()
		time.Sleep(time.Second)
	}
}
