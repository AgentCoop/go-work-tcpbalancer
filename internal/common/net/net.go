package net

import (
	n "net"
	"sync"
)

func NewConnManager(network string, addr string) *connManager {
	m := &connManager{network: network, addr: addr}
	m.inboundConns = make(ActiveConnectionsMap)
	m.outboundConns = make(ActiveConnectionsMap)
	m.newInbound = make(chan *Event, 100)
	m.newOutbound = make(chan *Event, 100)
	m.dataFrame = make(chan *Event, 100)
	m.outboundClosed = make(chan *Event)

	m.lisMap = make(listenAddrMap)
	return m
}

func (m *connManager) NewActiveConn(conn n.Conn, typ ConnType) *ActiveConn {
	ac := &ActiveConn{conn: conn, typ: typ}
	ac.eventMap = make(EventMap)
	ac.eventMapMu = new(sync.Mutex)

	ac.writeChan = make(chan interface{})
	ac.writeDoneChan = make(chan int)
	ac.readChan = make(chan interface{})


	ac.onNewConnChan = make(chan struct{}, 1)
	ac.onConnCloseChan = make(chan struct{}, 1)
	ac.onDataFrameChan = make(chan []byte)
	ac.onRawDataChan = make(chan []byte)

	ac.connManager = m
	ac.df = NewDataFrame()
	ac.readbuf = make([]byte, StreamReadBufferSize)
	return ac
}

func (m *connManager) GetBytesSent() uint64 {
	return m.bytesSent
}

func (m *connManager) GetBytesReceived() uint64 {
	return m.bytesReceived
}

func (m *connManager) NewInboundEvent() chan *Event {
	return m.newInbound
}

func (m *connManager) NewOutboundEvent() chan *Event {
	return m.newOutbound
}

func (m *connManager) OutboundClosedEvent() chan *Event {
	return m.outboundClosed
}

func (m *connManager) DataFrameEvent() chan *Event {
	return m.dataFrame
}

func (m *connManager) RawDataEvent() chan *Event {
	return m.rawdata
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
