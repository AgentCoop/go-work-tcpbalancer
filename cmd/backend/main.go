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
	cruncherPort, imgPort := backend.CliOptions.CruncherPort, backend.CliOptions.ImgResizePort

	if cruncherPort == 0 && imgPort == 0 {
		os.Exit(-1)
	}

	if backend.CliOptions.CruncherPort > 0 {
		connManager := newConnManager(backend.CliOptions.CruncherPort)
		go startCruncherServer(connManager)
		fmt.Printf("💻 [ %s:cruncher ] is listening on port %d\n",
			backend.CliOptions.Name, backend.CliOptions.CruncherPort)
	}

	if backend.CliOptions.ImgResizePort > 0 {
		connManager := newConnManager(backend.CliOptions.ImgResizePort)
		go startImgServer(connManager)
		fmt.Printf("💻 [ %s:img ] is listening on port %d\n",
			backend.CliOptions.Name, backend.CliOptions.ImgResizePort)
	}

	for {
		time.Sleep(time.Millisecond)
	}
}
