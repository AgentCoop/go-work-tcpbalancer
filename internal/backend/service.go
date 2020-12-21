package backend

import (
	"fmt"
	job "github.com/AgentCoop/go-work"
	"github.com/AgentCoop/go-work-tcpbalancer/internal/common/net"
	"strings"
)

func EchoService(j job.Job) (func(), func() interface{}, func()) {
	run := func() interface{} {
		cm := j.GetValue().(net.ConnManager)
		for key, v := range cm.GetInboundConns() {
			select {
			case data := <-v.GetReadChan():
				v.GetWriteChan() <- []byte(fmt.Sprintf("%s [%s] -> pong:  %s\n",
					key, CliOptions.Name, strings.ToUpper(string(data))))
			default:
			}
		}
		return nil
	}
	return nil, run, nil
}
