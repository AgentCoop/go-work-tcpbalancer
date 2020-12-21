#!/usr/bin/env sh

go mod tidy -v
go build -o /opt/client ./cmd/frontend/*.go

/opt/client -h proxy-server:9090
