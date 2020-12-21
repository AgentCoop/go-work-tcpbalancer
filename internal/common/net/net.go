package net

import (
	job "github.com/AgentCoop/go-work"
	"net"
	"sync/atomic"
)

func NewConnManager(network string, addr string) *connManager {
	m := &connManager{network: network, addr: addr}
	m.inboundConns = make(InboundConnections)
	return m
}

func (m *connManager) SetDataHandler(h IncomingDataHandler) {
	m.inHandler = h
}

func (m *connManager) GetInboundConns() InboundConnections {
	m.inboundConnMu.RLock()
	defer m.inboundConnMu.RUnlock()
	return m.inboundConns
}

func (m *connManager) GetBytesSent() uint64 {
	return m.bytesSent
}

func (m *connManager) GetBytesReceived() uint64 {
	return m.bytesReceived
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

func (c *inboundConn) GetReadChan() chan []byte {
	return c.readChan
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

		//pconn.SetDeadline(time.Now().Add(6 * time.Second))

		go func() {
			in := &inboundConn{conn: pconn}
			in.writeChan = make(chan []byte, 1)
			in.readChan = make(chan []byte, 1)
			in.connManager = cm
			in.connManager.inboundConnMu.Lock()
			in.connManager.inboundConns[in.Key()] = in
			in.connManager.inboundConnMu.Unlock()
			inJob := job.NewJob(in)
			inJob.AddTask(inboundConnWriteTask)
			inJob.AddTask(inboundConnReadTask)
			<-inJob.Run()
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
		in.readChan <- buf[0:n]
		return nil
	}
	cancel := func() {
		in := j.GetValue().(*inboundConn)
		cm := in.connManager
		atomic.AddInt32(&cm.inboundCounter, -1)
		in.connManager.inboundConnMu.Lock()
		delete(in.connManager.inboundConns, in.Key())
		in.connManager.inboundConnMu.Unlock()
		in.conn.Close()
	}
	return init, run, cancel
}
