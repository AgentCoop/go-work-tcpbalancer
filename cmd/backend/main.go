package main

import (
	"fmt"
	j "github.com/AgentCoop/go-work"
	"github.com/AgentCoop/go-work-tcpbalancer/internal/backend"
	"github.com/AgentCoop/go-work-tcpbalancer/internal/backend/task"
	n "github.com/AgentCoop/go-work-tcpbalancer/internal/common/net"
	"os"
	"strconv"
	"time"
)

var mainJob j.Job

func newConnManager(port int) n.ConnManager {
	localAddr := backend.CliOptions.Name + ":" + strconv.Itoa(port)
	connManager := n.NewConnManager("tcp4", localAddr)
	return connManager
}

func startCruncherServer(connManager n.ConnManager) {
	for {
		mainJob := j.NewJob(nil)
		mainJob.AddOneshotTask(connManager.AcceptTask)
		mainJob.AddTask(connManager.ReadTask)
		mainJob.AddTask(connManager.WriteTask)
		mainJob.AddTask(backend.CruncherTask)
		<-mainJob.RunInBackground()
		fmt.Printf("done job\n")
	}
}

func startImgServer(connManager n.ConnManager) {
	for {
		mainJob := j.NewJob(nil)
		mainJob.AddOneshotTask(connManager.AcceptTask)
		mainJob.AddTask(connManager.ReadTask)
		mainJob.AddTask(connManager.WriteTask)
		mainJob.AddTask(task.ResizeImageTask)
		<-mainJob.RunInBackground()
	}
}

func main() {
	backend.ParseCliOptions()

	if backend.CliOptions.Port == 0 {
		fmt.Printf("Specify a TCP port to listen on\n")
		os.Exit(-1)
	}

	connManager := newConnManager(backend.CliOptions.Port)

	switch backend.CliOptions.Service {
	case "cruncher":
		go startCruncherServer(connManager)
		fmt.Printf("💻 [ %s:cruncher ] is listening on port %d\n",
			backend.CliOptions.Name, backend.CliOptions.Port)
	case "img":
		go startImgServer(connManager)
		fmt.Printf("💻 [ %s:img ] is listening on port %d\n",
			backend.CliOptions.Name, backend.CliOptions.Port)
	}

	for {
		time.Sleep(time.Second)
	}
}
