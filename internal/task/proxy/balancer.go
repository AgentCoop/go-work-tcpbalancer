package proxy

import (
	"fmt"
	job "github.com/AgentCoop/go-work"
	netmanager "github.com/AgentCoop/net-manager"
	"math/rand"
	"net"
	"time"
)


type UpstreamList []*netmanager.ServerNet

type Balancer struct {
	upstreamServers UpstreamList
}

func NewBalancer() *Balancer {
	b := &Balancer{}
	//b.upstreamServers = make(UpstreamList)
	return b
}

func (b *Balancer) AddServer(item *netmanager.ServerNet) {
	b.upstreamServers = append(b.upstreamServers, item)
}

func (b *Balancer) SelectRandom() *netmanager.ServerNet {
	idx := rand.Intn(len(b.upstreamServers))
	upstreamSrv := b.upstreamServers[idx]
	return upstreamSrv
}

type proxy struct {
	conn netmanager.ProxyConn
}

func (p *proxy) downstream(j job.JobInterface) (job.Init, job.Run, job.Finalize) {
	init := func(task *job.TaskInfo) {
		p.conn.Upstream().DataKind = netmanager.DataRawKind
	}
	run := func(task *job.TaskInfo) {
		select {
		case dd := <- p.conn.Downstream().RecvRaw():
			p.conn.Upstream().Write() <- dd
			p.conn.Upstream().WriteSync()
			p.conn.Downstream().RecvRawSync()
			task.Tick()
		default:
			task.Idle()
		}
	}
	return init, run, func(task *job.TaskInfo) {
		p.conn.Upstream().CloseWithReuse()
	}
}

func (p *proxy) upstream(j job.JobInterface) (job.Init, job.Run, job.Finalize) {
	init := func(task *job.TaskInfo) {
		p.conn.Downstream().DataKind = netmanager.DataRawKind
	}
	run := func(task *job.TaskInfo) {
		select {
		case data := <- p.conn.Upstream().RecvRaw():
			p.conn.Downstream().Write() <- data
			p.conn.Downstream().WriteSync()
			p.conn.Upstream().RecvRawSync()
		}
		task.Tick()
	}
	return init, run, func(task *job.TaskInfo) {
		p.conn.Downstream().Close()
		j.Log(1) <- fmt.Sprintf("close proxy conn, downstream")
	}
}

func (p *proxy) ReadUpstreamTask(j job.JobInterface) (job.Init, job.Run, job.Finalize) {
	u := p.conn.Upstream()
	return u.ReadOnStreamTask(j)
}

func (b Balancer) LoadBalance(j job.JobInterface) (job.Init, job.Run, job.Finalize) {
	// Resolve hostnames of the upstream servers
	init := func(task *job.TaskInfo) {
		for _, srv := range b.upstreamServers {
			tcpAddr, err := net.ResolveTCPAddr("tcp4", srv.Server.Host)
			task.Assert(err)
			srv.TcpAddr = tcpAddr
		}
	}
	run := func(task *job.TaskInfo) {
		clientConn := j.GetValue().(*netmanager.StreamConn)
		select {
		case <- clientConn.NewConn():
			j.Log(0) <- fmt.Sprintf("new conn from")
			connMngr := clientConn.GetConnManager()
			netMngr := connMngr.GetNetworkManager()

			upstreamSrv := b.SelectRandom()
			proxyConn := netMngr.NewProxyConn(upstreamSrv, clientConn)
			p := &proxy{conn: proxyConn}

			pjob := job.NewJob(upstreamSrv.TcpAddr)
			pjob.AddOneshotTask(proxyConn.ProxyConnectTask)
			pjob.AddTask(p.conn.ProxyReadDownstreamTask)
			pjob.AddTask(p.conn.ProxyReadUpstreamTask)
			pjob.AddTask(p.conn.ProxyWriteDownstreamTask)
			pjob.AddTask(p.conn.ProxyWriteUpstreamTask)
			pjob.AddTaskWithIdleTimeout(p.downstream, time.Second * 2) // client connection timeout
			pjob.AddTask(p.upstream)
			<-pjob.RunInBackground()

			go func() {
				for {
					select {
					case <- pjob.GetDoneChan():
						_, err := pjob.GetInterruptedBy()
						pjob.Log(0) <- fmt.Sprintf("proxy conn job is %s, error %s", pjob.GetState(), err)
						j.Finish()
						return
					}
				}
			}()
		}
	}
	return init, run, nil
}