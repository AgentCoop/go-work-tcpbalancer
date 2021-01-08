package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	//"runtime/pprof"
	job "github.com/AgentCoop/go-work"
	"github.com/AgentCoop/go-work-tcpbalancer/internal/task/proxy"
	netmanager "github.com/AgentCoop/net-manager"
	"strconv"
	"time"
)

func runLoadBalancer(connMngr netmanager.ConnManager) {
	balancer := proxy.NewBalancer()
	for _, host := range CliOptions.UpstreamServers {
		srv := &netmanager.ServerNet{}
		srv.Server = &netmanager.Server{}
		srv.Server.Host = host
		balancer.AddServer(srv)
	}

	for {
		job := job.NewJob(nil)
		job.AddOneshotTask(connMngr.AcceptTask)
		job.AddTask(balancer.LoadBalance)
		<-job.RunInBackground()

		go func() {
			for {
				select {
				case <- job.GetDoneChan():
					_, err := job.GetInterruptedBy()
					job.Log(0) <- fmt.Sprintf("job is %s, error %s",  job.GetState(), err)
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
	localAddr := "localhost:" + strconv.Itoa(CliOptions.Port)
	opts := &netmanager.ConnManagerOptions{}
	opts.ReadbufLen = 256_00
	opts.InboundLimit = CliOptions.MaxClientConns
	connMngr := netMngr.NewConnManager("tcp4", localAddr, opts)

	go runLoadBalancer(connMngr)

	go func() {
		log.Println(http.ListenAndServe("localhost:6062", nil))
	}()

	for {
		time.Sleep(time.Second)
	}
}
