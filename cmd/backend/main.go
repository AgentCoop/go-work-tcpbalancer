package main

import (
	"fmt"
	"github.com/AgentCoop/go-work"
	t "github.com/AgentCoop/go-work-tcpbalancer/internal/task/backend"
	"github.com/AgentCoop/net-manager"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"
)

var counter int

func startImgServer(connManager netmanager.ConnManager) {
	for {
		mainJob := job.NewJob(nil)
		mainJob.AddOneshotTask(connManager.AcceptTask)
		mainJob.AddTask(netmanager.ReadTask)
		mainJob.AddTask(netmanager.WriteTask)
		mainJob.AddTask(t.ResizeImageTask)
		<-mainJob.RunInBackground()
		go func() {
			j := mainJob
			for {
				select {
				case <-j.GetDoneChan():
					_, err := mainJob.GetInterruptedBy()
					j.Log(0) <- fmt.Sprintf("#%d job is %s, error: %s",
						counter + 1, strings.ToLower(j.GetState().String()), err)
					counter++
					return
				}
			}
		}()
	}
}

func main() {
	ParseCliOptions()

	if CliOptions.CpuProfile != "" {
		fmt.Println(CliOptions.CpuProfile)
		f, err := os.Create(CliOptions.CpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

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
	opts.ReadbufLen = 256_00
	connMngr := netMngr.NewConnManager("tcp4", localAddr, opts)


	go startImgServer(connMngr)
	fmt.Printf(" 💻[ %s:img ] is listening on port %d\n",
			CliOptions.Name, CliOptions.Port)

	go func() {
		log.Println(http.ListenAndServe("localhost:6061", nil))
	}()

	for {
		time.Sleep(time.Second)
	}
}
