package backend

import (
	"fmt"
	job "github.com/AgentCoop/go-work"
	"github.com/AgentCoop/go-work-tcpbalancer/internal/common/net"
	"math/rand"
	"strings"
	"time"
)

func EchoService(j job.Job) (func(), func() interface{}, func()) {
	run := func() interface{} {
		cm := j.GetValue().(net.ConnManager)
		cm.IterateOverConns(net.Inbound, func(c net.ActiveConn) {
			select {
			case data := <-c.GetReadChan():
				c.GetWriteChan() <- []byte(fmt.Sprintf("%s [%s] -> pong:  %s\n",
					c.GetConn().RemoteAddr(), CliOptions.Name, strings.ToUpper(string(data))))
			default:
			}
		})
		return nil
	}
	return nil, run, nil
}

func StressTestTask(j job.Job) (func(), func() interface{}, func()) {
	init := func() {
		rand.Seed(rand.Int63())
	}
	run := func() interface{} {
		cm := j.GetValue().(net.ConnManager)
		cm.IterateOverConns(net.Inbound, func(c net.ActiveConn) {
			select {
			case <-c.GetReadChan():
				go func() {
					min, max := CliOptions.RespMinTime, CliOptions.RespMaxTime
					t := time.Duration(rand.Intn(max - min) + min) * time.Millisecond
					time.Sleep(t)
					resDataLen := rand.Intn(CliOptions.RespDataMaxLen) + 1
					randData := make([]byte, resDataLen)
					rand.Read(randData)
					fmt.Printf("[%s] -> resp. time: %d ms; bytes: %d\n",
						CliOptions.Name,  t / time.Millisecond, resDataLen)
					c.GetWriteChan() <- randData
					//fmt.Printf("done\n")
					time.Sleep(1 * time.Millisecond)
					c.GetConn().Close()
				}()
			default:
			}
		})
		time.Sleep(time.Second)
		return nil
	}
	return init, run, nil
}
