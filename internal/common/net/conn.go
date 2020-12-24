package net

import (
	job "github.com/AgentCoop/go-work"
	n "net"
	"sync"
)

func (c *ActiveConn) String() string {
	return c.conn.RemoteAddr().String() + " -> " + c.conn.LocalAddr().String()
}

func (c *ActiveConn) Key() string {
	return c.String()
}

func (c *ActiveConn) GetConnManager() *connManager {
	return c.connManager
}

func (c *ActiveConn) GetOnNewConnChan() <-chan struct{} {
	return c.onNewConnChan
}

func(c *ActiveConn) SetValue(value interface{}) {
	c.ValueMu.Lock()
	defer c.ValueMu.Unlock()
	c.value = value
}

func(c *ActiveConn) GetValue() interface{} {
	c.ValueMu.RLock()
	defer c.ValueMu.RUnlock()
	return c.value
}

func (c *ActiveConn) GetOnDataFrameChan() <-chan []byte {
	return c.onDataFrameChan
}

func (c *ActiveConn) GetConn() n.Conn {
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

func (c *ActiveConn) GetEvent(typ EventType) *Event {
	c.eventMapMu.Lock()
	defer c.eventMapMu.Unlock()
	evt, ok := c.eventMap[typ]
	if ! ok {
		evt := &Event{
			typ:     typ,
			conn:    c,
			data:    nil,
			value:   nil,
			ValueMu: sync.RWMutex{},
		}
		c.eventMap[typ] = evt
		return evt
	}
	return evt
}
