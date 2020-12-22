package frontend

import (
	"fmt"
	job "github.com/AgentCoop/go-work"
	"github.com/AgentCoop/go-work-tcpbalancer/internal/common/net"
	"math/rand"
	"time"
)

func ClientReqTask(j job.Job) (func(), func() interface{}, func()) {
	var buf []byte
	init := func() {
		l := rand.Intn(CliOptions.ReqDataMaxLen) + 1
		buf = make([]byte, l)
		rand.Seed(rand.Int63())
		rand.Read(buf)
	}
	run := func() interface{} {
		cm := j.GetValue().(net.ConnManager)
		cm.IterateOverConns(net.Outbound, func(c net.ActiveConn) {
			c.GetWriteChan() <- buf
		})
		return true
	}
	return init, run, nil
}


func ClientRespTask(j job.Job) (func(), func() interface{}, func()) {
	var reqStartTime time.Time
	init := func() {
		reqStartTime = time.Now()
	}
	run := func() interface{} {
		cm := j.GetValue().(net.ConnManager)
		cm.IterateOverConns(net.Outbound, func(c net.ActiveConn) {
			select {
			case res := <- c.GetReadChan():
				elapsed := time.Now().Sub(reqStartTime)
				fmt.Printf("[%s] -> res. time: %d ms; bytes: %d\n",
					c.GetConn().RemoteAddr(), elapsed / time.Millisecond, len(res))
			default:
			}
		})
		time.Sleep(time.Microsecond)
		return nil
	}
	return init, run, nil
}
