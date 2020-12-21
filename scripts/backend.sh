#!/usr/bin/env sh

go mod tidy -v
go build -o /opt/backend ./cmd/backend/*.go

/opt/backend -p 9090 --stress-test --name=$SERVNAME
