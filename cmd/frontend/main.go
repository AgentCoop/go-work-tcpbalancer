package main

import (
	"encoding/gob"
	"fmt"
	j "github.com/AgentCoop/go-work"
	"github.com/AgentCoop/net-manager"
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"

	//	"github.com/AgentCoop/go-work-tcpbalancer/internal/frontend/task"
	"github.com/AgentCoop/go-work-tcpbalancer/internal/task/frontend"
	"os"
	//	"time"
)

func startCruncherClient(mngr netmanager.ConnManager) {
	cruncher := frontend.NewCruncher(CruncherOpts.MinBatchesPerConn, CruncherOpts.MaxBatchesPerConn,
		CruncherOpts.MinItemsPerBatch, CruncherOpts.MaxBatchesPerConn)

	mainJob := j.NewJob(nil)
	mainJob.AddOneshotTask(mngr.ConnectTask)
	mainJob.AddTask(mngr.ReadTask)
	mainJob.AddTask(mngr.WriteTask)
	mainJob.AddTask(cruncher.SquareNumsInBatchTask)
	<-mainJob.Run()
}

func resizeImages(mngr netmanager.ConnManager) {
	for i := 0; i < ImgResizeOpts.Times; i++ {
		imgResizer := frontend.NewImageResizer(ImgResizeOpts.ImgDir, ImgResizeOpts.OutputDir,
			ImgResizeOpts.Width, ImgResizeOpts.Height)

		mainJob := j.NewJob(nil)
		mainJob.AddOneshotTask(mngr.ConnectTask)
		mainJob.AddTask(mngr.ReadTask)
		mainJob.AddTask(mngr.WriteTask)
		mainJob.AddTask(imgResizer.ScanForImagesTask)
		mainJob.AddTask(imgResizer.SaveResizedImageTask)
		<-mainJob.Run()

		counter++
		fmt.Printf("run %d\n", counter)
		if mainJob.IsCancelled() {
			fmt.Printf("job failed %s\n", mainJob.GetState())
			os.Exit(-1)
		}
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

	if len(MainOptions.ProxyHost) == 0 {
		fmt.Printf("Specify a proxy server to connect to\n")
		os.Exit(-1)
	}

	gob.Register(&frontend.CruncherPayload{})
	netMngr := netmanager.NewNetworkManager()
	connMngr := netMngr.NewConnManager("tcp4", MainOptions.ProxyHost)

	j.DefaultLogLevel = MainOptions.LogLevel
	j.RegisterDefaultLogger(func() j.LogLevelMap {
		m := make(j.LogLevelMap)
		handler := func(record interface{}, level int) {
			fmt.Printf("%s\n", record.(string))
		}
		m[1] = j.NewLogLevelMapItem(make(chan interface{}), handler)
		m[2] = j.NewLogLevelMapItem(make(chan interface{}), handler)
		return m
	})

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
