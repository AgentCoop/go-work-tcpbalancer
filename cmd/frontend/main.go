package main

import (
	"encoding/gob"
	"fmt"
	j "github.com/AgentCoop/go-work"
	n "github.com/AgentCoop/go-work-tcpbalancer/internal/common/net"
	"github.com/AgentCoop/go-work-tcpbalancer/internal/frontend"
	"os"
	"time"
)

func connToProxy(connManager n.ConnManager) {
	mainJob := j.NewJob(connManager)
	//mainJob.WithPrerequisites(connManager.Connect(mainJob))
	time.Sleep(time.Millisecond)
	mainJob.AddOneshotTask(connManager.ConnectTask)
	mainJob.AddTask(connManager.ReadTask)
	mainJob.AddTask(connManager.WriteTask)
	mainJob.AddTask(frontend.SquareNumsInBatchTask)
	fmt.Printf("Wait for run\n")
	<-mainJob.Run()
	fmt.Printf("Done\n")
	//go func(){
	//	select {
	//	case err := <- mainJob.GetError():
	//		fmt.Printf("err %s\n", err)
	//	}
	//}()
}

func main() {
	frontend.ParseCliOptions()

	fmt.Printf("Host: %s\n", frontend.MainOptions.ProxyHost)
	if len(frontend.MainOptions.ProxyHost) == 0 {
		fmt.Printf("Specify a proxy server to connect to\n")
		os.Exit(-1)
	}

	gob.Register(&frontend.CruncherPayload{})
	connManager := n.NewConnManager("tcp4", frontend.MainOptions.ProxyHost)


	//mainJob.WithPrerequisites(connManager.Connect(mainJob))
	for {
		mainJob := j.NewJob(nil)
		mainJob.AddOneshotTask(connManager.ConnectTask)
		mainJob.AddTask(connManager.ReadTask)
		mainJob.AddTask(connManager.WriteTask)
		mainJob.AddTask(frontend.SquareNumsInBatchTask)
		fmt.Printf("Wait for run\n")
		//select {
		<-mainJob.Run()
		time.Sleep(500 * time.Millisecond)
	}
}
