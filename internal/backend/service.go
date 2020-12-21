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

func StressTestTask(j job.Job) (func(), func() interface{}, func()) {
	init := func() {
		rand.Seed(rand.Int63())
	}
	run := func() interface{} {
		cm := j.GetValue().(net.ConnManager)
		for _, v := range cm.GetInboundConns() {
			select {
			case <-v.GetReadChan():
				currIn := v
				go func() {
					min, max := CliOptions.RespMinTime, CliOptions.RespMaxTime
					t := time.Duration(rand.Intn(max - min) + min) * time.Millisecond
					time.Sleep(t)
					resDataLen := rand.Intn(CliOptions.RespDataMaxLen) + 1
					randData := make([]byte, resDataLen)
					rand.Read(randData)
					fmt.Printf("[%s] -> resp. time: %d ms; bytes: %d\n",
						CliOptions.Name,  t / time.Millisecond, resDataLen)
					currIn.GetWriteChan() <- randData
					currIn.GetConn().Close()
				}()
			default:
			}
		}
		return nil
	}
	return init, run, nil
}
