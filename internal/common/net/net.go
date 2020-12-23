package net

import (
	job "github.com/AgentCoop/go-work"
	"net"
	"sync"
)

func NewConnManager(network string, addr string) *connManager {
	m := &connManager{network: network, addr: addr}
	m.inboundConns = make(ActiveConnectionsMap)
	m.outboundConns = make(ActiveConnectionsMap)
	m.eventChan = make(chan *Event)
	m.newInbound = make(chan *Event)
	m.newOutbound = make(chan *Event)
	m.dataFrame = make(chan *Event)
	return m
}

//func (m *connManager) SetDataHandler(h IncomingDataHandler) {
//	m.inHandler = h
//}

func (m *connManager) IterateOverConns(typ ConnType, f func(c *ActiveConn)) int {
	var conns ActiveConnectionsMap
	var l *sync.RWMutex
	switch typ  {
	case Inbound:
		l = &m.inboundConnMu
		conns = m.inboundConns
	case Outbound:
		l = &m.outboundConnMu
		conns = m.outboundConns
	}
	l.RLock()
	l.RUnlock()
	//defer l.RUnlock()
	count := 0
	for _, c := range conns {
		//l.RUnlock()
		f(c)
		count++
		//l.RLock()
	}
	return count
}

func (m *connManager) GetBytesSent() uint64 {
	return m.bytesSent
}

func (m *connManager) GetBytesReceived() uint64 {
	return m.bytesReceived
}

func (m *connManager) GetEventChan() chan *Event {
	return m.eventChan
}

func (m *connManager) NewInboundEvent() chan *Event {
	return m.newInbound
}

func (m *connManager) NewOutboundEvent() chan *Event {
	return m.newOutbound
}

func (m *connManager) DataFrameEvent() chan *Event {
	return m.dataFrame
}

func (m *connManager) addConn(c *ActiveConn) {
	var l *sync.RWMutex
	var connMap ActiveConnectionsMap
	switch c.typ {
	case Inbound:
		l = &m.inboundConnMu
		connMap = m.inboundConns
	case Outbound:
		l = &m.outboundConnMu
		connMap = m.outboundConns
	}
	l.Lock()
	defer l.Unlock()
	connMap[c.Key()] = c
}

func (m *connManager) delConn(c *ActiveConn) {
	var l *sync.RWMutex
	var connMap ActiveConnectionsMap
	switch c.typ {
	case Inbound:
		l = &m.inboundConnMu
		connMap = m.inboundConns
	case Outbound:
		l = &m.outboundConnMu
		connMap = m.outboundConns
	}
	l.Lock()
	defer l.Unlock()
	delete(connMap, c.Key())
	c.conn.Close()
}

func (c *ActiveConn) String() string {
	return c.conn.RemoteAddr().String() + " -> " + c.conn.LocalAddr().String()
}

func (c *ActiveConn) Key() string {
	return c.String()
}

func (c *ActiveConn) GetConn() net.Conn {
	return c.conn
}

func (c *ActiveConn) GetNetJob() job.Job {
	return c.netJob
}

func (c *ActiveConn) GetReadChan() chan interface{} {
	return c.readChan
}

func (c *ActiveConn) GetWriteChan() chan<- interface{} {
	return c.writeChan
}

func (c *connManager) Connect(j job.Job) <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		conn, err := net.Dial(c.network, c.addr)
		if err != nil {
			j.CancelWithError(err)
			return
		}
		ac := &ActiveConn{conn: conn, typ: Outbound, state: Active}
		ac.writeChan = make(chan interface{}, 1)
		ac.readChan = make(chan interface{}, 1)
		ac.connManager = c
		newJob := job.NewJob(ac)
		newJob.AddTask(connReadTask)
		newJob.AddTask(connWriteTask)
		ac.netJob = j
		c.addConn(ac)
		newJob.Run()
		ch <- struct{}{}
		c.newOutbound <- &Event{ conn: ac }
	}()
	return ch
}

