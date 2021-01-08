package main

import (
	"encoding/gob"
	"fmt"
	"github.com/AgentCoop/go-work"
	"github.com/AgentCoop/go-work-tcpbalancer/internal/task/frontend"
	"github.com/AgentCoop/net-manager"
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"sync"
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

func startCruncherClient(mngr netmanager.ConnManager) {
	//cruncher := frontend.NewCruncher(CruncherOpts.MinBatchesPerConn, CruncherOpts.MaxBatchesPerConn,
	//	CruncherOpts.MinItemsPerBatch, CruncherOpts.MaxBatchesPerConn)
	//f := func() job.Job {
	//	mainJob := job.NewJob(nil)
	//	mainJob.AddOneshotTask(mngr.ConnectTask)
	//	mainJob.AddTask(netmanager.ReadTask)
	//	mainJob.AddTask(netmanager.WriteTask)
	//	mainJob.AddTask(cruncher.SquareNumsInBatchTask)
	//	return mainJob
	//}
}

func resizeImages(mngr netmanager.ConnManager) {
	n := ImgResizeOpts.Times / MainOptions.MaxConns
	for i := 0; i < n; i++ {
		f := func() job.Job {
			imgResizer := frontend.NewImageResizer(ImgResizeOpts.ImgDir, ImgResizeOpts.OutputDir,
				ImgResizeOpts.Width, ImgResizeOpts.Height)
			j := job.NewJob(nil)
			j.AddOneshotTask(mngr.ConnectTask)
			j.AddTask(netmanager.ReadTask)
			j.AddTask(netmanager.WriteTask)
			j.AddTask(imgResizer.ScanForImagesTask)
			j.AddTask(imgResizer.SaveResizedImageTask)
			return j
		}
		execInParallel(f, MainOptions.MaxConns)
	}
}

//func showNetStatistics(manager n.ConnManager) {
//	fmt.Printf("-- [ Network Statistics ] --\n")
//	fmt.Printf("\tbytes sent: %0.2f Mb\n", float64(manager.GetBytesSent()) / 1e6)
//	fmt.Printf("\tbytes received: %0.2f Mb\n", float64(manager.GetBytesReceived()) / 1e6)
//}
var counter int

func main() {
	ParseCliOptions()
	initLogger()

	gob.Register(&frontend.CruncherPayload{})
	netMngr := netmanager.NewNetworkManager()
	connMngr := netMngr.NewConnManager("tcp4", MainOptions.ProxyHost, nil)

	runtime.SetBlockProfileRate(6)
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	switch MainOptions.Service {
	case "cruncher":
		startCruncherClient(connMngr)
	case "imgresize":
		resizeImages(connMngr)
	}

	//showNetStatistics(connManager)
}
