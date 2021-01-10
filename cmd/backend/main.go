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
		//mainJob.AddTaskWithIdleTimeout(netmanager.ReadTask, time.Second * 12)
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
	ParseCliOptions()

	//if CliOptions.CpuProfile != "" {
	//	fmt.Println(CliOptions.CpuProfile)
	//	f, err := os.Create(CliOptions.CpuProfile)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	runtime.SetCPUProfileRate(100)
	//	pprof.StartCPUProfile(f)
	//	defer pprof.StopCPUProfile()
	//}

	// Set up logger
	job.DefaultLogLevel = CliOptions.LogLevel
	job.RegisterDefaultLogger(func() job.LogLevelMap {
		m := make(job.LogLevelMap, 3)
		handler := func(record interface{}, level int) {
			prefix := fmt.Sprintf(" 💻[ %s ] ", CliOptions.Name) +
				strings.Repeat("☞ ", level)
			fmt.Printf("%s%s\n", prefix, record.(string))
		}
		m[0] = job.NewLogLevelMapItem(make(chan interface{}), handler)
		m[1] = job.NewLogLevelMapItem(make(chan interface{}), handler)
		m[2] = job.NewLogLevelMapItem(make(chan interface{}), handler)
		return m
	})

	netMngr := netmanager.NewNetworkManager()
	localAddr := CliOptions.Name + ":" + strconv.Itoa(CliOptions.Port)
	opts := &netmanager.ConnManagerOptions{}
	opts.ReadbufLen = 4096
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
