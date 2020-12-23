package frontend

import (
	"bytes"
	"encoding/gob"
	"fmt"
	job "github.com/AgentCoop/go-work"
	"github.com/AgentCoop/go-work-tcpbalancer/internal/common/net"
	"time"
)

type CruncherPayload struct {
	ItemsCount int
	Items []uint32
}

type CruncherResult struct {
	SquaredNums []uint64
}

func SquareNumsInBatchTask(j job.Job) (func(), func() interface{}, func()) {
	var reqStartTime time.Time
	init := func() {
		reqStartTime = time.Now()
	}
	run := func() interface{} {
		cm := j.GetValue().(net.ConnManager)
		select {
		case evt := <- cm.NewOutboundEvent():
			fmt.Printf("New conn to\n")
			count := 6
			items := make([]uint32, count)
			for i, v := range [...]uint32{1,2,3,4,5,6} {
				items[i] = v
			}
			c := evt.GetActiveConn()
			payload := &CruncherPayload{Items: items, ItemsCount: count}
			c.DataFrame = payload
			fmt.Printf(" <- req items %d\n", count)
			c.GetWriteChan() <- payload
		case evt := <- cm.DataFrameEvent():
			buf := bytes.NewBuffer(evt.GetData())
			dec := gob.NewDecoder(buf)
			nums := &CruncherResult{}
			err := dec.Decode(nums)
			j.Assert(err)
			fmt.Printf("Got crunched numbers %v\n", nums.SquaredNums)
		}
		return true
	}
	return init, run, nil
}
