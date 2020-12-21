package net

import (
	job "github.com/AgentCoop/go-work"
	n "net"
	"sync"
)

type ConnType int

const (
	Inbound = iota
	Outbound
)

func (s ConnType) String() string {
	return [...]string{"Inbound", "Outbound"}[s]
}

type IncomingDataHandler func(data []byte, c ActiveConn)
type ActiveConnectionsMap map[string]*activeConn

type ConnManager interface {
	SetDataHandler(h IncomingDataHandler)
	GetInboundConns() ActiveConnectionsMap
	GetOutboundConns() ActiveConnectionsMap
	GetBytesSent() uint64
	GetBytesReceived() uint64
}

type ActiveConn interface {
	GetWriteChan() chan<- []byte
	GetReadChan() <-chan []byte
	GetConn() n.Conn
	GetNetJob() job.Job
}

type activeConn struct {
	conn n.Conn
	writeChan chan []byte
	readChan chan []byte
	connManager *connManager
	netJob	job.Job
}

type connManager struct {
	typ             ConnType
	inboundCounter  int32
	outboundCounter int32
	bytesSent       uint64
	bytesReceived   uint64
	inHandler       IncomingDataHandler
	inboundConnMu   sync.RWMutex
	inboundConns    ActiveConnectionsMap
	outboundConnMu  sync.RWMutex
	outboundConns   ActiveConnectionsMap
	plis            n.Listener
	network         string
	addr            string
}
