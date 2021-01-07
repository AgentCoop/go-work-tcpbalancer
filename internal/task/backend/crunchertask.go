package backend

import (
	"fmt"
	"github.com/AgentCoop/go-work"
	"github.com/AgentCoop/go-work-tcpbalancer/internal/task/frontend"
	"github.com/AgentCoop/net-manager"
)

func crunchNumbers(payload *frontend.CruncherPayload, stream netmanager.Stream) {
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
	stream.Write() <- result
	stream.WriteSync()
}

func CruncherTask(j job.Job) (job.Init, job.Run, job.Finalize) {
	run := func(t job.Task) {
		ac := j.GetValue().(netmanager.Stream)
		for {
			select {
			//case <-ac.GetOnNewConnChan():
			case frame := <-ac.RecvDataFrame():
				payload := &frontend.CruncherPayload{}
				err := frame.Decode(payload)
				fmt.Printf(" <- new numbers to crunch %d\n", payload.ItemsCount)
				t.Assert(err)
				go crunchNumbers(payload, ac)
			}
		}
	}
	return nil, run, nil
}
