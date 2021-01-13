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

func (p *proxy) downstream(j job.Job) (job.Init, job.Run, job.Finalize) {
	run := func(task job.Task) {
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
	fin := func(task job.Task) {
		p.conn.Downstream().Close()
	}
	return nil, run, fin
}

func (p *proxy) upstream(j job.Job) (job.Init, job.Run, job.Finalize) {
	run := func(task job.Task) {
		select {
		case data := <- p.conn.Upstream().RecvRaw(): // Receive data from upstream server
			p.conn.Downstream().Write() <- data // Write data to downstream server
			p.conn.Downstream().WriteSync() // sync with downstream data receiver
			p.conn.Upstream().RecvRawSync() // sync with upstream data sender
		}
		task.Tick()
	}
	fin := func(task job.Task) {
		p.conn.Upstream().CloseWithReuse()
	}
	return nil, run, fin
}

func (p *proxy) ReadUpstreamTask(j job.Job) (job.Init, job.Run, job.Finalize) {
	u := p.conn.Upstream()
	return u.ReadOnStreamTask(j)
}

func (b Balancer) LoadBalance(j job.Job) (job.Init, job.Run, job.Finalize) {
	// Resolve hostnames of the upstream servers
	init := func(task job.Task) {
		for _, srv := range b.upstreamServers {
			tcpAddr, err := net.ResolveTCPAddr("tcp4", srv.Server.Host)
			task.Assert(err)
			srv.TcpAddr = tcpAddr
		}
	}
	run := func(task job.Task) {
		clientConn := j.GetValue().(netmanager.Stream)
		select {
		case <- clientConn.NewConn():
			j.Log(1) <- fmt.Sprintf("conn from %s", clientConn.String())

			upstreamSrv := b.SelectRandom()
			proxyConn := netmanager.NewProxyConn(upstreamSrv, clientConn)
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

			select {
			case <- pjob.GetDoneChan():
				_, err := pjob.GetInterruptedBy()
				pjob.Log(2) <- fmt.Sprintf("proxy conn job is %s, error %s", pjob.GetState(), err)
				j.Finish()
				return
			}
			task.Done()
		default:
			task.Tick()
		}
	}
	return init, run, nil
}