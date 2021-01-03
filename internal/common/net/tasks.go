package net

import (
	"fmt"
	job "github.com/AgentCoop/go-work"
	"net"
	"sync/atomic"
)

func (c *connManager) ConnectTask(j job.JobInterface) (job.Init, job.Run, job.Cancel) {
	run := func(t *job.TaskInfo) {
		conn, err := net.Dial(c.network, c.addr)
		t.Assert(err)
		ac := c.NewActiveConn(conn, Outbound)
		j.SetValue(ac)
		c.addConn(ac)
		//t.GetDoneChan() <- job.DoneSig
		ac.onNewConnChan <- job.DoneSig
		t.Done()
	}
	return nil, run, nil
}

func (c *connManager) AcceptTask(j job.JobInterface) (job.Init, job.Run, job.Cancel) {
	run := func(t *job.TaskInfo) {
		var lis net.Listener
		key := c.network + c.addr
		if _, ok := c.lisMap[key]; ! ok {
			l, err := net.Listen(c.network, c.addr)
			c.lisMap[key] = l
			t.Assert(err)
			lis = l
		}
		lis = c.lisMap[key]

		conn, acceptErr := lis.Accept()
		t.Assert(acceptErr)

		atomic.AddInt32(&c.inboundCounter, 1)
		//pconn.SetDeadline(time.Now().Add(6 * time.Second))
		ac := c.NewActiveConn(conn, Inbound)
		j.SetValue(ac)
		c.addConn(ac)

		ac.onNewConnChan <- job.DoneSig
		t.Done()
	}
	cancel := func() {
		fmt.Printf("Reader Task finishes\n")
	}
	return nil, run, cancel
}

func (c *connManager) WriteTask(j job.JobInterface) (job.Init, job.Run, job.Cancel) {
	run := func(t *job.TaskInfo) {
		//fmt.Printf("run write\n")
		ac := j.GetValue().(*ActiveConn)
		var n int
		var err error
		select {
		case data := <- ac.writeChan:
			switch data.(type) {
			case []byte: // raw data
				n, err = ac.conn.Write(data.([]byte))
				t.Assert(err)
			case nil:
				j.Log(2) <- fmt.Sprintf(" -> nil data")
				// Handle error
			default:
				enc, err := ac.df.ToFrame(data)
				t.Assert(err)
				n, err = ac.conn.Write(enc)
				t.Assert(err)
			}
			// Sync with the writer
			ac.writeDoneChan <- n
			j.Log(2) <- fmt.Sprintf(" -> bytes %d wrote", n)
			atomic.AddUint64(&ac.connManager.bytesSent, uint64(n))
		}
		t.Tick()
	}
	cancel := func()  {
		ac := j.GetValue().(*ActiveConn)
		close(ac.writeChan)
		close(ac.writeDoneChan)
		ac.conn.Close()
	}
	return nil, run, cancel
}

func (c *connManager) ReadTask(j job.JobInterface) (job.Init, job.Run, job.Cancel) {
	run := func(t *job.TaskInfo) {
		ac := j.GetValue().(*ActiveConn)
		n, err := ac.conn.Read(ac.readbuf)
		t.Assert(err)
		atomic.AddUint64(&ac.connManager.bytesReceived, uint64(n))
		ac.df.Capture(ac.readbuf[0:n])

		if ac.df.IsFullFrame() {
			f := ac.df.GetFrame()
			ac.onDataFrameChan <- f
			<-ac.OnDataFrameDoneChan
		}

		t.Tick()
	}
	cancel := func() {
		ac := j.GetValue().(*ActiveConn)
		cm := ac.connManager
		atomic.AddInt32(&cm.outboundCounter, -1)
		close(ac.readChan)
		close(ac.onDataFrameChan)
		close(ac.onRawDataChan)
		cm.delConn(ac)
		ac.conn.Close()
	}
	return nil, run, cancel
}
