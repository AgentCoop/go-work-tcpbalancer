package main

import (
	"fmt"
	j "github.com/AgentCoop/go-work"
	"github.com/AgentCoop/go-work-tcpbalancer/internal/backend"
	n "github.com/AgentCoop/go-work-tcpbalancer/internal/common/net"
	t "github.com/AgentCoop/go-work-tcpbalancer/internal/task/backend"
	"github.com/AgentCoop/net-manager"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"time"
)

var mainJob j.Job

func newConnManager(port int) n.ConnManager {
	localAddr := backend.CliOptions.Name + ":" + strconv.Itoa(port)
	connManager := n.NewConnManager("tcp4", localAddr)
	return connManager
}

func startCruncherServer(connManager netmanager.ConnManager) {
	for {
		mainJob := j.NewJob(nil)
		mainJob.AddOneshotTask(connManager.AcceptTask)
		mainJob.AddTask(connManager.ReadTask)
		mainJob.AddTask(connManager.WriteTask)
		mainJob.AddTask(t.CruncherTask)
		<-mainJob.RunInBackground()
		fmt.Printf("done job\n")
	}
}

func startImgServer(connManager netmanager.ConnManager) {
	for {
		mainJob := j.NewJob(nil)
		mainJob.AddOneshotTask(connManager.AcceptTask)
		mainJob.AddTask(connManager.ReadTask)
		mainJob.AddTask(connManager.WriteTask)
		mainJob.AddTask(t.ResizeImageTask)
		<-mainJob.RunInBackground()
		go func() {
			for {
				select {
				case <-mainJob.GetDoneChan():
					fmt.Printf("Job done\n")
					return
				}
			}
		}()
		fmt.Printf("num goroutines %d\n", runtime.NumGoroutine())
	}
}

func main() {
	backend.ParseCliOptions()

	if backend.CliOptions.CpuProfile != "" {
		fmt.Println(backend.CliOptions.CpuProfile)
		f, err := os.Create(backend.CliOptions.CpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if backend.CliOptions.Port == 0 {
		fmt.Printf("Specify a TCP port to listen on\n")
		os.Exit(-1)
	}

	netMngr := netmanager.NewNetworkManager()
	localAddr := backend.CliOptions.Name + ":" + strconv.Itoa(backend.CliOptions.Port)
	connMngr := netMngr.NewConnManager("tcp4", localAddr)

	switch backend.CliOptions.Service {
	case "cruncher":
		go startCruncherServer(connMngr)
		fmt.Printf("💻 [ %s:cruncher ] is listening on port %d\n",
			backend.CliOptions.Name, backend.CliOptions.Port)
	case "img":
		go startImgServer(connMngr)
		fmt.Printf("💻 [ %s:img ] is listening on port %d\n",
			backend.CliOptions.Name, backend.CliOptions.Port)
	}

	go func() {
		log.Println(http.ListenAndServe("localhost:6061", nil))
	}()

	for {
		time.Sleep(time.Second)
	}
}
