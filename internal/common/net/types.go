package net

import (
	job "github.com/AgentCoop/go-work"
	"github.com/AgentCoop/net-dataframe"
	n "net"
	"sync"
)

type ConnType int
type ConnState int
type EventType int
type EventMap map[EventType]*Event
type listenAddrMap map[string]n.Listener

const (
	Inbound ConnType = iota
	Outbound
)

const (
	Active ConnState = iota
	Closed
)

const (
	DataFrame EventType = iota
	RawStream
	NewOutboundConn
	NewInboundConn
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

	ConnectTask(j job.JobInterface) (job.Init, job.Run, job.Cancel)
	AcceptTask(j job.JobInterface) (job.Init, job.Run, job.Cancel)
	ReadTask(j job.JobInterface) (job.Init, job.Run, job.Cancel)
	WriteTask(j job.JobInterface) (job.Init, job.Run, job.Cancel)

	DataFrameEvent() chan *Event
	RawDataEvent() chan *Event
	NewInboundEvent() chan *Event
	NewOutboundEvent() chan *Event
	OutboundClosedEvent() chan *Event
}

const (
	StreamReadBufferSize = 4096
)

type Request struct {
	Size uint64
	Body interface{}
}

type ActiveConn struct {
	conn n.Conn

	writeChan chan interface{}
	writeDoneChan chan int

	readChan chan interface{}
	onNewConnChan chan struct{}
	onConnCloseChan chan struct{}

	onDataFrameChan chan []byte
	OnDataFrameDoneChan chan struct{}

	onRawDataChan chan []byte

	connManager *connManager
	netJob	job.Job
	state ConnState
	typ ConnType
	eventMapMu	*sync.Mutex
	eventMap EventMap
	df *netdataframe.DataFrame
	readbuf []byte

	value   interface{}
	ValueMu sync.RWMutex
}

type Event struct {
	typ     EventType
	conn    *ActiveConn
	data    []byte
	value   interface{}
	ValueMu sync.RWMutex
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
	outboundClosed chan *Event
	dataFrame   chan *Event
	rawdata     chan *Event

	lisMap listenAddrMap
	network       string
	addr          string
}
