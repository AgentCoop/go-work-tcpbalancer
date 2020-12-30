package frontend

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	job "github.com/AgentCoop/go-work"
	"github.com/AgentCoop/go-work-tcpbalancer/internal/common/net"
	"math/rand"
)

type Cruncher struct {
	minItems uint // per batch
	maxItems uint
	minBatches uint // per connection
	maxBatches uint
}

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

// Returns a random integer number from the specified range
func randInt(min int, max int) int {
	if max <= min {
		max = min + 1
	}
	return rand.Intn(max - min + 1) + min
}

func NewCruncher(minBatches uint, maxBatches uint, minItems uint, maxItems uint) *Cruncher {
	c := &Cruncher{
		minItems:   1,
		maxItems:   10,
		minBatches: 1,
		maxBatches: 1,
	}
	if minBatches > 0 { c.minBatches = minBatches }
	if maxBatches > 0 { c.maxBatches = maxBatches }
	if minItems > 0 { c.minItems = minItems }
	if maxItems > 0 { c.maxItems = maxItems }
	return c
}

func (c *Cruncher) newBatch(num int) *CruncherPayload {
	count := randInt(int(c.minItems), int(c.maxItems))
	r := &CruncherPayload{}
	r.ItemsCount = count
	r.Items = make([]uint32,count)
	r.BatchNum = num
	for i := 0; i < count; i++ {
		r.Items[i] = uint32(randInt(10, 100))
	}
	return r
}

func (c *Cruncher) dispatchBatch(ac *net.ActiveConn) {
	nBatches := randInt(int(c.minBatches), int(c.maxBatches))
	// Map request data with response
	bp := make(BatchMap)
	for i := 0; i < nBatches; i++{
		batch := c.newBatch(i + 1)
		bp[i + 1] = batch
	}
	ac.SetValue(bp)
	for _, v := range bp {
		fmt.Printf(" ->  batch #%d, items %d\n", v.BatchNum, v.ItemsCount)
		ac.GetWriteChan() <- v
	}
}

func (c *Cruncher) SquareNumsInBatchTask(j job.JobInterface) (job.Init, job.Run, job.Cancel) {
	run := func(t *job.TaskInfo) {
		ac := j.GetValue().(*net.ActiveConn)
		cm := ac.GetConnManager()
		//fmt.Printf("Connected\n")
		select {
		case <-ac.GetOnNewConnChan():
			go c.dispatchBatch(ac)
		case raw := <- cm.RawDataEvent():
			fmt.Printf("Raw data %v\n", raw)
		case frame := <- ac.GetOnDataFrameChan():
			buf := bytes.NewBuffer(frame)
			dec := gob.NewDecoder(buf)
			nums := &CruncherResult{}
			err := dec.Decode(nums)
			j.Assert(err)

			//fmt.Printf("Got crunched numbers for batch #%d, %d\n", nums.BatchNum, len(nums.SquaredNums))

			batchMap := ac.GetValue().(BatchMap)
			if batchMap == nil {
				fmt.Printf("Empty batch map\n")
				return
			}

			if nums == nil {
				fmt.Printf("Empty payload")
				return
			}

			batch, ok := batchMap[nums.BatchNum]
			if  ! ok {
				fmt.Printf("No batch\n")
				return
			}

			ac.ValueMu.Lock()
			for i := 0; i < batch.ItemsCount; i++ {
				if uint64(batch.Items[i] * batch.Items[i]) != nums.SquaredNums[i] {
					j.Assert(errors.New("Batch processing failed"))
				}
				delete(batchMap, nums.BatchNum)
			}
			fmt.Printf(" -> batch #%d with %d items verified\n", nums.BatchNum, len(nums.SquaredNums))
			ac.ValueMu.Unlock()

			if len(batchMap) == 0 {
				// Close current connection, no more batches to dispatch
				j.Cancel()
				return
			}
		}
		t.Done()
	}

	cancel := func() {
		fmt.Printf("Canceling job\n")
	}

	return nil, run, cancel
}
