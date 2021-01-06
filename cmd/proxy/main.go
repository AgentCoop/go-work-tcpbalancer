package main

import (
	"fmt"
	//"fmt"
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
	connMngr := netMngr.NewConnManager("tcp4", localAddr)

	go runLoadBalancer(connMngr)

	for {
		time.Sleep(time.Second)
	}
}
