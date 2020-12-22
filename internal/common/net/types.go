package net

import (
	job "github.com/AgentCoop/go-work"
	n "net"
	"sync"
)

type ConnType int
type ConnState int

const (
	Inbound ConnType = iota
	Outbound
)

const (
	Active ConnState = iota
	Closed
)

func (s ConnType) String() string {
	return [...]string{"Inbound", "Outbound"}[s]
}

func (s ConnState) String() string {
	return [...]string{"Active", "Closed"}[s]
}

type IncomingDataHandler func(data []byte, c ActiveConn)
type ActiveConnectionsMap map[string]*activeConn

type ConnManager interface {
	SetDataHandler(h IncomingDataHandler)
	IterateOverConns(typ ConnType, f func(c ActiveConn)) int
	GetBytesSent() uint64
	GetBytesReceived() uint64
	Connect(j job.Job) <-chan struct{}
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
	state ConnState
	typ ConnType
}

type connManager struct {
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
