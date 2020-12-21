package net

import (
	job "github.com/AgentCoop/go-work"
	"net"
	"sync/atomic"
	"time"
)

func NewConnManager(network string, addr string) *connManager {
	m := &connManager{network: network, addr: addr}
	return m
}

func (m *connManager) SetDataHandler(h IncomingDataHandler) {
	m.inHandler = h
}

func (m *connManager) GetInboundConns() InboundConnections {
	return m.inboundConns
}

func (c *inboundConn) String() string {
	return c.conn.RemoteAddr().String() + " -> " + c.conn.LocalAddr().String()
}

func (c *inboundConn) Key() string {
	return c.String()
}

func (c *inboundConn) GetConn() net.Conn {
	return c.conn
}

func (c *inboundConn) GetWriteChan() chan<- []byte {
	return c.writeChan
}

func ListenTask(j job.Job) (func(), func() interface{}, func()) {
	init := func() {
		cm := j.GetValue().(*connManager)
		lis, err := net.Listen(cm.network, cm.addr)
		cm.plis = lis
		j.Assert(err)
	}

	run := func() interface{} {
		cm := j.GetValue().(*connManager)
		pconn, acceptErr := cm.plis.Accept()
		j.Assert(acceptErr)
		atomic.AddInt32(&cm.inboundCounter, 1)

		pconn.SetDeadline(time.Now().Add(6 * time.Second))

		go func() {
			ac := &inboundConn{conn: pconn}
			ac.writeChan = make(chan []byte)
			ac.connManager = cm
			acJob := job.NewJob(ac)
			acJob.AddTask(inboundConnWriteTask)
			acJob.AddTask(inboundConnReadTask)
			<-acJob.Run()
		}()
		return nil
	}

	cancel := func() { }
	return init, run, cancel
}

func inboundConnWriteTask(j job.Job) (func(), func() interface{}, func()) {
	var buf []byte
	init := func() {
		buf = make([]byte, 1024)
	}
	run := func() interface{} {
		ac := j.GetValue().(*inboundConn)
		select {
		case data := <- ac.writeChan:
			n, err := ac.conn.Write(data)
			atomic.AddUint64(&ac.connManager.bytesSent, uint64(n))
			j.Assert(err)
		}
		return nil
	}
	return init, run, nil
}

func inboundConnReadTask(j job.Job) (func(), func() interface{}, func()) {
	var buf []byte
	init := func() {
		buf = make([]byte, 1024)
	}
	run := func() interface{} {
		in := j.GetValue().(*inboundConn)
		n, err := in.conn.Read(buf)
		atomic.AddUint64(&in.connManager.bytesReceived, uint64(n))
		j.Assert(err)
		if in.connManager.inHandler != nil {
			in.connManager.inHandler(buf[0:n], in)
		}
		return nil
	}
	cancel := func() {
		in := j.GetValue().(*inboundConn)
		cm := in.connManager
		atomic.AddInt32(&cm.inboundCounter, -1)
		in.conn.Close()
	}
	return init, run, cancel
}
