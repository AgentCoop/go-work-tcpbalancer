package main

import (
	"fmt"
	"github.com/AgentCoop/go-work"
	"github.com/AgentCoop/go-work-tcpbalancer/internal/task/frontend"
	"github.com/AgentCoop/net-manager"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"sync"
	"time"
)

var jobCounter int

// Executes the given job in parallel using N goroutines
func execInParallel(f func() job.Job, N int) {
	var wg sync.WaitGroup
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func() {
			job := f()
			select {
			case <- job.Run():
				job.Log(0) <- fmt.Sprintf("#%d job is %s", jobCounter+1, job.GetState())
				jobCounter++
				wg.Done()
				return
			}
		}()
	}
	wg.Wait()
}

func resizeImages(mngr netmanager.ConnManager) {
	nConns := int(rand.Int31n(int32(MainOptions.MaxConns))) + 1
	for i := 0; i < MainOptions.Times; i++ {
		f := func() job.Job {
			imgResizer := frontend.NewImageResizer(MainOptions.ImgDir, MainOptions.OutputDir,
				MainOptions.Width, MainOptions.Height)
			j := job.NewJob(nil)
			j.AddOneshotTask(mngr.ConnectTask)
			j.AddTask(netmanager.ReadTask)
			j.AddTask(netmanager.WriteTask)
			j.AddTask(imgResizer.ScanForImagesTask)
			j.AddTaskWithIdleTimeout(imgResizer.SaveResizedImageTask, time.Second * 1)
			return j
		}
		execInParallel(f, nConns)
	}
}

//func showNetStatistics(manager n.ConnManager) {
//	fmt.Printf("-- [ Network Statistics ] --\n")
//	fmt.Printf("\tbytes sent: %0.2f Mb\n", float64(manager.GetBytesSent()) / 1e6)
//	fmt.Printf("\tbytes received: %0.2f Mb\n", float64(manager.GetBytesReceived()) / 1e6)
//}
//var counter int

func main() {
	ParseCliOptions()
	initLogger()

	netMngr := netmanager.NewNetworkManager()
	connMngr := netMngr.NewConnManager("tcp4", MainOptions.ProxyHost, nil)

	runtime.SetBlockProfileRate(6)
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	resizeImages(connMngr)

	//showNetStatistics(connManager)
}
