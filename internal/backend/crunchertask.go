package backend

import (
	"bytes"
	"encoding/gob"
	"fmt"
	job "github.com/AgentCoop/go-work"
	"github.com/AgentCoop/go-work-tcpbalancer/internal/common/net"
	"github.com/AgentCoop/go-work-tcpbalancer/internal/frontend"
)

func crunchNumbers(payload *frontend.CruncherPayload, evt *net.Event) {
	result := &frontend.CruncherResult{}
	result.SquaredNums = make([]uint64, payload.ItemsCount)
	for i, num := range payload.Items {
		result.SquaredNums[i] = uint64(num * num)
	}
	// Send result back
	// If connection to this time was closed the goroutine will try to write to the closed channel as well
	// causing it panic and exit.
	c := evt.GetActiveConn()
	c.GetWriteChan() <- result
}

func StressTestTask(j job.Job) (func(), func() interface{}, func()) {
	init := func() {

	}
	run := func() interface{} {
		cm := j.GetValue().(net.ConnManager)
		for {
			select {
			case evt := <-cm.DataFrameEvent():
				buf := bytes.NewBuffer(evt.GetData())
				dec := gob.NewDecoder(buf)
				payload := &frontend.CruncherPayload{}
				err := dec.Decode(payload)
				go crunchNumbers(payload, evt)
				j.Assert(err)
				fmt.Printf("got evnt %v\n", evt)
			}
		}
		return true
	}
	return init, run, nil
}
