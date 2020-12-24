package net

import (
	"fmt"
	job "github.com/AgentCoop/go-work"
	"net"
	"sync/atomic"
)

func (c *connManager) ConnectTask(j job.Job) (func(), func() interface{}, func()) {
	run := func() interface{} {
		conn, err := net.Dial(c.network, c.addr)
		j.Assert(err)
		ac := c.NewActiveConn(conn, Outbound)
		j.SetValue(ac)
		c.addConn(ac)
		ac.onNewConnChan <- struct{}{}
		return true
	}
	return nil, run, nil
}

func (c *connManager) AcceptTask(j job.Job) (func(), func() interface{}, func()) {
	run := func() interface{} {
		var lis net.Listener
		key := c.network + c.addr
		if _, ok := c.lisMap[key]; ! ok {
			l, err := net.Listen(c.network, c.addr)
			c.lisMap[key] = l
			j.Assert(err)
			lis = l
		}
		lis = c.lisMap[key]

		conn, acceptErr := lis.Accept()
		j.Assert(acceptErr)

		atomic.AddInt32(&c.inboundCounter, 1)
		//pconn.SetDeadline(time.Now().Add(6 * time.Second))
		ac := c.NewActiveConn(conn, Inbound)
		j.SetValue(ac)
		c.addConn(ac)
		//evt := ac.GetEvent(NewInboundConn)
		ac.onNewConnChan <- struct{}{}
		return true
	}
	cancel := func() {
		fmt.Printf("Reader Task finishes\n")
	}
	return nil, run, cancel
}

func (c *connManager) WriteTask(j job.Job) (func(), func() interface{}, func()) {
	run := func() interface{} {
		ac := j.GetValue().(*ActiveConn)
		var n int
		var err error
		select {
		case data := <- ac.writeChan:
			switch data.(type) {
			case []byte: // raw data
				n, err = ac.conn.Write(data.([]byte))
				j.Assert(err)
			case nil:
				a := 1
				a++
				fmt.Printf("Got NIL data\n")
				// Handle error
			default:
				enc, err := ac.df.Encode(data)
				j.Assert(err)
				n, err = ac.conn.Write(enc)
				j.Assert(err)
			}
			atomic.AddUint64(&ac.connManager.bytesSent, uint64(n))
			j.Assert(err)
		}
		return nil
	}
	return nil, run, func() {
		fmt.Printf("Write Task finishes\n")
	}
}

func (c *connManager) ReadTask(j job.Job) (func(), func() interface{}, func()) {
	run := func() interface{} {
		ac := j.GetValue().(*ActiveConn)
		frame, err, raw := ac.df.Decode(ac.conn)
		j.Assert(err)
		if frame != nil {
			fmt.Printf("Reader send data frame\n")
			ac.onDataFrameChan <- frame
		} else if raw != nil {
			fmt.Printf("Reader send raw data\n")
			ac.onRawDataChan <- raw
		} else {
			fmt.Printf("Reader task no data\n")
		}
		return nil
	}
	cancel := func() {
		ac := j.GetValue().(*ActiveConn)
		cm := ac.connManager
		atomic.AddInt32(&cm.outboundCounter, -1)
		close(ac.readChan)
		close(ac.writeChan)
		cm.delConn(ac)
	}
	return nil, run, cancel
}
