package net

import (
	n "net"
)

type IncomingDataHandler func(data []byte, c InboundConn)
type InboundConnections map[string]*inboundConn

type ConnManager interface {
	SetDataHandler(h IncomingDataHandler)
	GetInboundConns() InboundConnections
	GetBytesSent() uint64
	GetBytesReceived() uint64
}

type InboundConn interface {
	GetWriteChan() chan<- []byte
	GetConn() n.Conn
}

type inboundConn struct {
	conn n.Conn
	writeChan chan []byte
	connManager *connManager
}

type connManager struct {
	inboundCounter		int32
	outboundCounter		int32
	bytesSent			uint64
	bytesReceived		uint64
	inHandler IncomingDataHandler
	inboundConns		InboundConnections
	plis n.Listener
	network string
	addr string
}
