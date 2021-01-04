package backend

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/AgentCoop/go-work"
	"github.com/AgentCoop/go-work-tcpbalancer/internal/common/net"
	"github.com/AgentCoop/go-work-tcpbalancer/internal/task/frontend"
)

func crunchNumbers(payload *frontend.CruncherPayload, ac *net.ActiveConn) {
	result := &frontend.CruncherResult{}
	result.SquaredNums = make([]uint64, payload.ItemsCount)
	result.BatchNum = payload.BatchNum
	for i, num := range payload.Items {
		result.SquaredNums[i] = uint64(num * num)
	}
	// Send result back
	// If connection to this time was closed the goroutine will try to write to the closed channel as well
	// causing it panic and exit.
	fmt.Printf(" <-send result back: batch #%d\n", result.BatchNum)
	ac.GetWriteChan() <- result
}

func CruncherTask(j job.JobInterface) (job.Init, job.Run, job.Cancel) {
	run := func(t *job.TaskInfo) {
		ac := j.GetValue().(*net.ActiveConn)
		for {
			select {
			case <-ac.GetOnNewConnChan():
			case frame := <-ac.GetOnDataFrameChan():
				buf := bytes.NewBuffer(frame)
				dec := gob.NewDecoder(buf)
				payload := &frontend.CruncherPayload{}
				err := dec.Decode(payload)
				fmt.Printf(" <- new numbers to crunch %d\n", payload.ItemsCount)
				t.Assert(err)
				go crunchNumbers(payload, ac)
			}
		}
	}
	return nil, run, nil
}
