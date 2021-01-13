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
		balancerJob := job.NewJob(nil)
		balancerJob.AddOneshotTask(connMngr.AcceptTask)
		balancerJob.AddTask(balancer.LoadBalance)
		<-balancerJob.RunInBackground()

		go func() {
			for {
				select {
				case <- balancerJob.GetDoneChan():
					_, err := balancerJob.GetInterruptedBy()
					balancerJob.Log(2) <- fmt.Sprintf("job is %s, error '%v'",  balancerJob.GetState(), err)
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
	localAddr := ":" + strconv.Itoa(CliOptions.Port)
	opts := &netmanager.ConnManagerOptions{}
	opts.ReadbufLen = 256_00
	opts.InboundLimit = CliOptions.MaxClientConns
	connMngr := netMngr.NewConnManager("tcp4", localAddr, opts)

	go runLoadBalancer(connMngr)
	fmt.Printf(" 🌎[ proxy server ] is listening on port %d\n", CliOptions.Port)

	if CliOptions.Debug {
		go func() {
			log.Println(http.ListenAndServe("localhost:6062", nil))
		}()
	}

	for {
		time.Sleep(time.Second)
	}
}
