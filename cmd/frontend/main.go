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

var (
	counterMu sync.RWMutex
	jobCounter int
	startedAt int64
	finishedAt int64
)

// Executes the given job in parallel using N goroutines
func execInParallel(f func() job.Job, N int32) {
	var wg sync.WaitGroup
	for i := 0; i < int(N); i++ {
		wg.Add(1)
		go func() {
			job := f()
			select {
			case <- job.Run():
				_, err := job.GetInterruptedBy()
				counterMu.RLock()
				job.Log(1) <- fmt.Sprintf("#%d job is %s, error '%v'", jobCounter+1, job.GetState(), err)
				counterMu.RUnlock()

				counterMu.Lock()
				jobCounter++
				counterMu.Unlock()

				wg.Done()
				return
			}
		}()
	}
	wg.Wait()
}

func resizeImages(mngr netmanager.ConnManager) {
	min, max := int32(cliOptions.MinConns), int32(cliOptions.MaxConns)
	var nConns int32
	for i := 0; i < cliOptions.Times; i++ {
		if max <= min {
			nConns = min
		} else {
			nConns = rand.Int31n(max) + min
		}
		f := func() job.Job {
			imgResizer := frontend.NewImageResizer(cliOptions.ImgDir, cliOptions.OutputDir,
				cliOptions.Width, cliOptions.Height, cliOptions.DryRun)
			j := job.NewJob(nil)
			j.AddOneshotTask(mngr.ConnectTask)
			j.AddTask(netmanager.ReadTask)
			j.AddTask(netmanager.WriteTask)
			j.AddTask(imgResizer.ScanForImagesTask)
			j.AddTaskWithIdleTimeout(imgResizer.SaveResizedImageTask, time.Second * 8)
			return j
		}
		execInParallel(f, nConns)
	}
}

func showNetStatistics(connMngr netmanager.ConnManager) {
	fmt.Printf("-- [ Network Statistics ] --\n")
	fmt.Printf("\tbytes sent: %0.4f Mb\n", float64(connMngr.GetBytesSent()) / 1e6)
	fmt.Printf("\tbytes received: %0.4f Mb\n", float64(connMngr.GetBytesReceived()) / 1e6)
	rps := float64(time.Duration(jobCounter) * time.Second) / float64(finishedAt - startedAt)
	fmt.Printf("\tRequests Per Second: %0.2f\n", rps)
}

func main() {
	ParseCliOptions()
	initLogger()

	netMngr := netmanager.NewNetworkManager()
	opts := &netmanager.ConnManagerOptions{}
	opts.ReadbufLen = 20000
	connMngr := netMngr.NewConnManager("tcp4", cliOptions.ProxyHost, opts)

	if cliOptions.Debug {
		go func() {
			runtime.SetBlockProfileRate(6)
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	startedAt = time.Now().UnixNano()
	resizeImages(connMngr)
	finishedAt = time.Now().UnixNano()

	//time.Sleep(time.Second * 6)
	showNetStatistics(connMngr)
}
