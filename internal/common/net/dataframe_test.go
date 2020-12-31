package net_test

import (
	"github.com/AgentCoop/go-work-tcpbalancer/internal/common/net"
	"math/rand"
	"testing"
)

const (
	MinDataFrameLen = 500
	MaxDataFrameLen = ^int32(0) >> 16
)

type payload struct {
	a int8
	bulkData []byte
	b int16
	c uint64
	msg string
}

var (
	df = net.NewDataFrame()
)

func transferDataFrame(a int8, bulkData []byte, b int16, c uint64, msg string) *payload {
	p := &payload{a, bulkData, b, c,msg}
	df.

	return p
}

func TestConnState_String(t *testing.T) {
	var head int
	for i := 0; i < 100; i++ {
		datalen := MinDataFrameLen + rand.Int31n(MaxDataFrameLen - MinDataFrameLen)
		data := make([]byte, datalen)
		_, err := rand.Read(data)
		if err != nil { panic(err) }
		for head = 0 ; head < len(data);
	}
}