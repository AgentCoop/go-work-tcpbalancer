package net

import (
	"fmt"
	job "github.com/AgentCoop/go-work"
	"net"
	"sync/atomic"
)

func NewConnManager(network string, addr string) *connManager {
	m := &connManager{network: network, addr: addr}
	m.inboundConns = make(ActiveConnectionsMap)
	m.outboundConns = make(ActiveConnectionsMap)
	return m
}

func (m *connManager) SetDataHandler(h IncomingDataHandler) {
	m.inHandler = h
}

func (m *connManager) GetInboundConns() ActiveConnectionsMap {
	m.inboundConnMu.RLock()
	defer m.inboundConnMu.RUnlock()
	return m.inboundConns
}

func (m *connManager) GetOutboundConns() ActiveConnectionsMap {
	m.outboundConnMu.RLock()
	defer m.outboundConnMu.RUnlock()
	return m.outboundConns
}

func (m *connManager) GetBytesSent() uint64 {
	return m.bytesSent
}

func (m *connManager) GetBytesReceived() uint64 {
	return m.bytesReceived
}

func (m *connManager) addOutboundConn(c *activeConn) {
	m.outboundConnMu.Lock()
	defer m.outboundConnMu.Unlock()
	m.outboundConns[c.Key()] = c
}

func (c *activeConn) String() string {
	return c.conn.RemoteAddr().String() + " -> " + c.conn.LocalAddr().String()
}

func (c *activeConn) Key() string {
	return c.String()
}

func (c *activeConn) GetConn() net.Conn {
	return c.conn
}

func (c *activeConn) GetNetJob() job.Job {
	return c.netJob
}

func (c *activeConn) GetReadChan() chan []byte {
	return c.readChan
}

func (c *activeConn) GetWriteChan() chan<- []byte {
	return c.writeChan
}

func (c *connManager) Connect(j job.Job) chan struct{} {
	ch := make(chan struct{})
	go func() {
		conn, err := net.Dial(c.network, c.addr)
		if err != nil {
			fmt.Printf("conn err %s\n", err)
			j.CancelWithError(err)
			return
		}
		fmt.Println("Connected")
		ac := &activeConn{conn: conn}
		ac.writeChan = make(chan []byte, 1)
		ac.readChan = make(chan []byte, 1)
		ac.connManager = c
		newJob := job.NewJob(ac)
		newJob.AddTask(connReadTask)
		newJob.AddTask(connWriteTask)
		ac.netJob = newJob
		c.addOutboundConn(ac)
		newJob.Run()
		fmt.Println("Run new job")
		ch <- struct{}{}
	}()
	return ch
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
			in := &activeConn{conn: pconn}
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
		ac := j.GetValue().(*activeConn)
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
		in := j.GetValue().(*activeConn)
		n, err := in.conn.Read(buf)
		atomic.AddUint64(&in.connManager.bytesReceived, uint64(n))
		j.Assert(err)
		in.readChan <- buf[0:n]
		return nil
	}
	cancel := func() {
		in := j.GetValue().(*activeConn)
		cm := in.connManager
		atomic.AddInt32(&cm.inboundCounter, -1)
		in.connManager.inboundConnMu.Lock()
		delete(in.connManager.inboundConns, in.Key())
		in.connManager.inboundConnMu.Unlock()
		in.conn.Close()
	}
	return init, run, cancel
}

func connWriteTask(j job.Job) (func(), func() interface{}, func()) {
	var buf []byte
	init := func() {
		buf = make([]byte, 1024)
	}
	run := func() interface{} {
		ac := j.GetValue().(*activeConn)
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

func connReadTask(j job.Job) (func(), func() interface{}, func()) {
	var buf []byte
	init := func() {
		buf = make([]byte, 1024)
	}
	run := func() interface{} {
		ac := j.GetValue().(*activeConn)
		n, err := ac.conn.Read(buf)
		j.Assert(err)
		atomic.AddUint64(&ac.connManager.bytesReceived, uint64(n))
		ac.readChan <- buf[0:n]
		return nil
	}
	cancel := func() {
		ac := j.GetValue().(*activeConn)
		cm := ac.connManager
		atomic.AddInt32(&cm.outboundCounter, -1)
		cm.outboundConnMu.Lock()
		delete(cm.outboundConns, ac.Key())
		cm.outboundConnMu.Unlock()
		fmt.Printf(" client conn %s closed\n",ac.GetConn().RemoteAddr())
		ac.conn.Close()
		fmt.Printf("Done\n")
	}
	return init, run, cancel
}