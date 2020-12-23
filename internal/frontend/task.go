package frontend

import (
	"bytes"
	"encoding/gob"
	"fmt"
	job "github.com/AgentCoop/go-work"
	"github.com/AgentCoop/go-work-tcpbalancer/internal/common/net"
	"math/rand"
	"time"
)

type CruncherPayload struct {
	BatchNum int
	ItemsCount int
	Items []uint32
}

type BatchMap map[int]*CruncherPayload

type CruncherResult struct {
	BatchNum int
	SquaredNums []uint64
}

func randInt(min int, max int) int {
	if max == min {
		return min
	} else {
		return rand.Intn(max - min + 1) + min
	}
}

func newBatch(num int) *CruncherPayload {
	min, max := CliOptions.MinItemsPerBatch, CliOptions.MaxItemsPerBatch
	count := randInt(min, max)
	r := &CruncherPayload{}
	r.ItemsCount = count
	r.Items = make([]uint32,count)
	r.BatchNum = num
	for i := 0; i < count; i++ {
		r.Items[i] = uint32(randInt(10, 100))
	}
	return r
}

func dispatchBatch(evt *net.Event) {
	min, max := CliOptions.MinBatchesPerConn, CliOptions.MaxBatchesPerConn
	nBatches := randInt(min, max)
	bp := make(BatchMap)
	for i := 0; i < nBatches; i++{
		batch := newBatch(i + 1)
		bp[i + 1] = batch
	}
	evt.SetValue(bp)
	for i, v := range bp {
		fmt.Printf(" ->  batch #%d, items %d\n", i + 1, v.ItemsCount)
		evt.GetActiveConn().GetWriteChan() <- v
	}
}

func SquareNumsInBatchTask(j job.Job) (func(), func() interface{}, func()) {
	var reqStartTime time.Time
	var nBatches, nProcessed int
	init := func() {
		reqStartTime = time.Now()
	}
	run := func() interface{} {
		cm := j.GetValue().(net.ConnManager)
		select {
		case evt := <- cm.NewOutboundEvent():
			go dispatchBatch(evt)
		case evt := <- cm.DataFrameEvent():
			buf := bytes.NewBuffer(evt.GetData())
			dec := gob.NewDecoder(buf)
			nums := &CruncherResult{}
			err := dec.Decode(nums)
			j.Assert(err)
			nProcessed++
			batch := evt.GetValue()
			fmt.Printf("%v\n", batch)
			//fmt.Printf("Got crunched numbers for batch #%d %v\n", batch.BatchNum, nums.SquaredNums)
			fmt.Printf("proceed %d\n", nProcessed)
			if (nProcessed == nBatches) {
				// Close current connection, no more batches to dispatch
				evt.GetActiveConn().GetNetJob().Cancel()
			}
		}
		return true
	}
	return init, run, nil
}
