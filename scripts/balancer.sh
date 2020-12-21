#!/usr/bin/env sh

go mod tidy -v
go build -o /opt/tcpbalancer ./cmd/balancer/*.go

/opt/tcpbalancer -p 9900 -u ieft.rg:80
