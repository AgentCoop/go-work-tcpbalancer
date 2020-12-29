package main

import (
	"encoding/gob"
	"fmt"
	j "github.com/AgentCoop/go-work"
	n "github.com/AgentCoop/go-work-tcpbalancer/internal/common/net"
	"github.com/AgentCoop/go-work-tcpbalancer/internal/frontend"
	"github.com/AgentCoop/go-work-tcpbalancer/internal/frontend/task"
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

func startCruncherClient(manager n.ConnManager) {
	mainJob := j.NewJob(nil)
	mainJob.AddOneshotTask(manager.ConnectTask)
	mainJob.AddTask(manager.ReadTask)
	mainJob.AddTask(manager.WriteTask)
	mainJob.AddTask(frontend.SquareNumsInBatchTask)
	<-mainJob.Run()
}

func startImgResizerClient(manager n.ConnManager) {
	mainJob := j.NewJob(nil)
	mainJob.AddOneshotTask(manager.ConnectTask)
	mainJob.AddTask(manager.ReadTask)
	mainJob.AddTask(manager.WriteTask)
	mainJob.AddTask(task.ScanForImagesTask)
	mainJob.AddTask(task.SaveResizedImageTask)
	<-mainJob.Run()
	fmt.Printf("-- [ Network Statistics ] --\n")
	fmt.Printf("\tbytes sent: %0.2f Mb\n", float64(manager.GetBytesSent() / 10e6))
	fmt.Printf("\tbytes received: %0.2f Mb\n", float64(manager.GetBytesReceived()))
}

func main() {
	frontend.ParseCliOptions()

	if len(frontend.MainOptions.ProxyHost) == 0 {
		fmt.Printf("Specify a proxy server to connect to\n")
		os.Exit(-1)
	}

	gob.Register(&frontend.CruncherPayload{})
	connManager := n.NewConnManager("tcp4", frontend.MainOptions.ProxyHost)

	switch frontend.MainOptions.Service {
	case "cruncher":
		startCruncherClient(connManager)
	case "imgresize":
		startImgResizerClient(connManager)
	}
}
