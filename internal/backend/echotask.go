package backend

import (
	job "github.com/AgentCoop/go-work"
	"github.com/AgentCoop/go-work-tcpbalancer/internal/common/net"
)

func EchoService(j job.Job) (func(), func() interface{}, func()) {
	run := func() interface{} {
		cm := j.GetValue().(net.ConnManager)
		cm.IterateOverConns(net.Inbound, func(c *net.ActiveConn) {
			//select {
			//case data := <-c.GetReadChan():
			//	c.GetWriteChan() <- []byte(fmt.Sprintf("%s [%s] -> pong:  %s\n",
			//		c.GetConn().RemoteAddr(), CliOptions.Name, strings.ToUpper(string(data))))
			//default:
			//}
		})
		return nil
	}
	return nil, run, nil
}
