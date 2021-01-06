package common

import "net"

type Server struct {
	Host string
	IpAddr 			net.IPAddr
	Weight uint8
	MaxConns uint16
}

type ServerNet struct {
	Server *Server
	TcpAddr *net.TCPAddr
}
