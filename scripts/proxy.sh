#!/usr/bin/env sh

go mod tidy -v
go build -o /opt/proxy ./cmd/proxy/*.go

/opt/proxy \
  -p 9090 \
  -u backend-1:9090 \
  -u backend-2:9090 \
  -u backend-3:9090 \
  --loglevel=2
