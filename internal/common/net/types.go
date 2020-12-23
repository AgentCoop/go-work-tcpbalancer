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

const (
	InDataFrame EventType = iota
	InRawStream
	OutbandConnEstablished
	InboundConnEstablished
)

func (s ConnType) String() string {
	return [...]string{"Inbound", "Outbound"}[s]
}

func (s ConnState) String() string {
	return [...]string{"Active", "Closed"}[s]
}

//type IncomingDataHandler func(data []byte, c ActiveConn)
type ActiveConnectionsMap map[string]*ActiveConn

type ConnManager interface {
	//SetDataHandler(h IncomingDataHandler)
	IterateOverConns(typ ConnType, f func(c *ActiveConn)) int
	GetEventChan() chan *Event
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
	Request *Request
	DataFrame interface{}
}

type Event struct {
	conn      *ActiveConn
	data      []byte
	eventType EventType
}

func(e *Event) GetEventType() EventType {
	return e.eventType
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
	//inHandler       IncomingDataHandler
	inboundConnMu   sync.RWMutex
	inboundConns    ActiveConnectionsMap
	outboundConnMu  sync.RWMutex
	outboundConns   ActiveConnectionsMap

	eventChan   chan *Event
	newInbound  chan *Event
	newOutbound chan *Event
	dataFrame   chan *Event
	rawstream    chan *Event

	plis            n.Listener
	network         string
	addr            string
}
