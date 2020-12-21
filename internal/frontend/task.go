package frontend

import (
	"fmt"
	job "github.com/AgentCoop/go-work"
	"github.com/AgentCoop/go-work-tcpbalancer/internal/common/net"
	"math/rand"
	"time"
)

func ClientTask(j job.Job) (func(), func() interface{}, func()) {
	var buf []byte
	init := func() {
		l := rand.Intn(CliOptions.ReqDataMaxLen) + 1
		buf = make([]byte, l)
		rand.Seed(rand.Int63())
		rand.Read(buf)
	}
	run := func() interface{} {
		cm := j.GetValue().(net.ConnManager)
		for _, v := range cm.GetOutboundConns() {
			fmt.Printf("Iterate over\n")
			v.GetWriteChan() <- buf
			start := time.Now()
			select {
			case res := <- v.GetReadChan():
				elapsed := time.Now().Sub(start)
				fmt.Printf("[%s] -> res. time: %d ms; bytes: %d\n",
					v.GetConn().RemoteAddr(), elapsed / time.Millisecond, len(res))
			}
			v.GetNetJob().Cancel()
		}
		return true
	}
	return init, run, nil
}
