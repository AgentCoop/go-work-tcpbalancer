package net

import (
	job "github.com/AgentCoop/go-work"
	n "net"
	"sync"
)

type ConnType int
type ConnState int
type EventType int

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

type ActiveConnectionsMap map[string]*ActiveConn

type ConnManager interface {
	GetBytesSent() uint64
	GetBytesReceived() uint64
	Connect(j job.Job) <-chan struct{}

	DataFrameEvent() chan *Event
	NewInboundEvent() chan *Event
	NewOutboundEvent() chan *Event
}

//type ActiveConn interface {
//	GetWriteChan() chan<- interface{}
//	GetReadChan() <-chan []byte
//	GetConn() n.conn
//	GetNetJob() job.Job
//}

const (
	StreamReadBufferSize = 4096
)

var dataFrameMagicWord = [...]byte{ 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88 }

type DataFrame struct {
	len uint64
	Data interface{}
	readbuf []byte
	tail []byte
}

type Request struct {
	Size uint64
	Body interface{}
}

type ActiveConn struct {
	conn n.Conn
	writeChan chan interface{}
	readChan chan interface{}
	connManager *connManager
	netJob	job.Job
	state ConnState
	typ ConnType
}

type Event struct {
	conn      *ActiveConn
	data      []byte
}

func(e *Event) GetData() []byte {
	return e.data
}

func(e *Event) GetActiveConn() *ActiveConn {
	return e.conn
}

type connManager struct {
	inboundCounter  int32
	outboundCounter int32
	bytesSent       uint64
	bytesReceived   uint64

	inboundConnMu   sync.RWMutex
	inboundConns    ActiveConnectionsMap
	outboundConnMu  sync.RWMutex
	outboundConns   ActiveConnectionsMap

	newInbound  chan *Event
	newOutbound chan *Event
	dataFrame   chan *Event
	rawstream    chan *Event

	plis            n.Listener
	network         string
	addr            string
}
