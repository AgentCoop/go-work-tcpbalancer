package frontend

import (
	"bytes"
	"encoding/gob"
	"fmt"
	job "github.com/AgentCoop/go-work"
	"github.com/AgentCoop/go-work-tcpbalancer/internal/common/net"
	"math/rand"
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

func dispatchBatch(ac *net.ActiveConn) {
	min, max := CliOptions.MinBatchesPerConn, CliOptions.MaxBatchesPerConn
	nBatches := randInt(min, max)

	// Map request data with response
	bp := make(BatchMap)
	for i := 0; i < nBatches; i++{
		batch := newBatch(i + 1)
		bp[i + 1] = batch
	}
	ac.SetValue(bp)

	for _, v := range bp {
		fmt.Printf(" ->  batch #%d, items %d\n", v.BatchNum, v.ItemsCount)
		ac.GetWriteChan() <- v
	}
}

func SquareNumsInBatchTask(j job.Job) (func(), func() interface{}, func()) {
	init := func() {	}

	run := func() interface{} {
		ac := j.GetValue().(*net.ActiveConn)
		cm := ac.GetConnManager()
		fmt.Printf("Connected\n")
		select {
		case <-ac.GetOnNewConnChan():
			fmt.Printf("New conn\n")
			go dispatchBatch(ac)
		case raw := <- cm.RawDataEvent():
			fmt.Printf("Raw data %v\n", raw)
		case frame := <- ac.GetOnDataFrameChan():
			fmt.Printf("New response frame\n")
			buf := bytes.NewBuffer(frame)
			dec := gob.NewDecoder(buf)
			nums := &CruncherResult{}
			err := dec.Decode(nums)
			j.Assert(err)

			fmt.Printf("Got crunched numbers for batch #%d %v\n", nums.BatchNum, nums.SquaredNums)

			batchMap := ac.GetValue().(BatchMap)
			if batchMap == nil {
				fmt.Printf("Empty batch map\n")
				return true
			}

			if nums == nil {
				fmt.Printf("Empty payload")
				return true
			}

			batch, ok := batchMap[nums.BatchNum]
			if  ! ok {
				fmt.Printf("No batch\n")
				return true
			}

			fmt.Printf("batch map %d, batch #%d: [%v]\n", len(batchMap), batch.BatchNum, batch.Items)
			//j.AssertTrue(ok, "failed to loop up batch")

			for i := 0; i < batch.ItemsCount; i++ {
				//if uint64(batch.Items[i] * batch.Items[i]) != nums.SquaredNums[i] {
				//	fmt.Printf("#2\n")
				//	j.Assert(errors.New("Batch processing failed"))
				//}
				//evt.ValueMu.Lock()
				delete(batchMap, nums.BatchNum)
				//evt.ValueMu.Unlock()
			}
			//return true
			if len(batchMap) == 0 {
				// Close current connection, no more batches to dispatch
				fmt.Printf("Finish batch crunching\n")
				j.Cancel()
				return true
			}
		}
		return nil
	}

	cancel := func() {
		fmt.Printf("Canceling job\n")
	}

	return init, run, cancel
}
