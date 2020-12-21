#!/usr/bin/env sh

go mod tidy -v
go build -o /opt/backend ./cmd/backend/*.go

/opt/backend -p 9900 --echo --name=$SERVNAME
