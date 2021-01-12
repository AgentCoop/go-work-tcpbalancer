package main

import (
	"fmt"
	"github.com/AgentCoop/go-work"
	t "github.com/AgentCoop/go-work-tcpbalancer/internal/task/backend"
	"github.com/AgentCoop/net-manager"
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var counter int

func startImgServer(connManager netmanager.ConnManager) {
	opts := &t.ResizerOptions{}
	for {
		fmt.Printf("new conn\n")
		mainJob := job.NewJob(opts)
		mainJob.AddOneshotTask(connManager.AcceptTask)
		mainJob.AddTask(netmanager.ReadTask)
		//mainJob.AddTaskWithIdleTimeout(netmanager.ReadTask, time.Second * 2)
		mainJob.AddTask(netmanager.WriteTask)
		mainJob.AddTask(opts.ResizeImageTask)
		<-mainJob.RunInBackground()
		go func() {
			j := mainJob
			for {
				select {
				case <-j.GetDoneChan():
					_, err := mainJob.GetInterruptedBy()
					j.Log(1) <- fmt.Sprintf("#%d job is %s, error: %s",
						counter + 1, strings.ToLower(j.GetState().String()), err)
					counter++
					j.Log(1) <- fmt.Sprintf("N gouroutines %d", runtime.NumGoroutine())
 					return
				}
			}
		}()
	}
}

func main() {
	parseCliOptions()
	initLogger()

	netMngr := netmanager.NewNetworkManager()
	localAddr := CliOptions.Name + ":" + strconv.Itoa(CliOptions.Port)
	opts := &netmanager.ConnManagerOptions{}
	opts.ReadbufLen = 60_000
	connMngr := netMngr.NewConnManager("tcp4", localAddr, opts)
	//_ = connMngr
	go startImgServer(connMngr)
	fmt.Printf(" 💻[ %s ] is listening on port %d\n",
			CliOptions.Name, CliOptions.Port)

	go func() {
		log.Println(http.ListenAndServe("localhost:6061", nil))
	}()

	for {
		time.Sleep(time.Second)
	}
}
