#!/usr/bin/env sh

go mod tidy -v
go build -o /opt/tcpbalancer ./cmd/balancer/*.go

/opt/tcpbalancer \
  -p 9090 \
  -u backend-1:9090 \
  -u backend-2:9090 \
  -u backend-3:9090
