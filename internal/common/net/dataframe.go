package net

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	n "net"
)

func NewDataFrame() *dataFrame {
	f := &dataFrame{}
	f.tail = nil
	return f
}

func (f *dataFrame) Encode(data interface{}) ([]byte, error) {
	var frame bytes.Buffer
	// Encode frame data
	enc := gob.NewEncoder(&frame)
	err := enc.Encode(data)
	if err != nil { return nil, err }

	// Write frame length in big endian order
	framelen := uint64(frame.Len())
	lbuf := new(bytes.Buffer)
	err = binary.Write(lbuf, binary.BigEndian, framelen)
	if err != nil { return nil, err }

	// Compose data stream
	mwl := len(dataFrameMagicWord)
	buf := make([]byte, mwl + lbuf.Len() + frame.Len())
	copy(buf[0:mwl], dataFrameMagicWord[:]) // data frame magic word
	copy(buf[mwl:], lbuf.Bytes()) // length
	copy(buf[mwl+lbuf.Len():], frame.Bytes()) // serialized data structure
	return buf, nil
}

func (f *dataFrame) Decode(conn n.Conn) ([]byte, error, []byte) {
	var framebuf []byte
	if f.readbuf == nil {
		f.readbuf = make([]byte, StreamReadBufferSize)
	}

	n1, err := conn.Read(f.readbuf)
	if err != nil || n1 == 0 { return nil, err, nil }

	if len(f.tail) > 0 { // Prepend tail data of the previous frame
		switch len(f.tail)+n1 > cap(f.readbuf) {
		case true:
			newbuf := make([]byte, len(f.tail)+n1)
			copy(newbuf[len(f.tail):n1], f.readbuf[0:n1])
			f.readbuf = newbuf
		default:
			copy(f.readbuf[len(f.tail):n1], f.readbuf[0:n1])
		}
		copy(f.readbuf[0:], f.tail)
		f.tail = f.tail[:0]
	}

	if f.probe() {
		mwl := len(dataFrameMagicWord)
		r := bytes.NewReader(f.readbuf[mwl:2*mwl])
		var fl uint64
		var nn int
		binary.Read(r, binary.BigEndian, &fl)
		framebuf = make([]byte, fl)
		tail := n1-8-mwl
		for copy(framebuf[0:tail], f.readbuf[mwl+8:n1]); tail < len(framebuf); tail += nn {
			nn, err = r.Read(f.readbuf)
			if err != nil { return nil, err, nil }
			if tail+nn <= len(framebuf) {
				copy(framebuf[tail:nn], f.readbuf[:nn])
			} else {
				left := tail + nn - cap(framebuf)
				copy(framebuf[tail:left], f.readbuf[0:left])
				f.tail = make([]byte, nn-left)
				copy(f.tail, f.readbuf[left:nn])
			}
		}
		return framebuf, nil, nil
	} else {
		return nil, nil, f.readbuf[0:n1]
	}
}

func (f *dataFrame) probe() bool {
	if len(f.readbuf) < len(dataFrameMagicWord) {
		return false
	}
	for i := 0; i < len(dataFrameMagicWord); i++ {
		if f.readbuf[i] != dataFrameMagicWord[i] {
			return false
		}
	}
	return true
}
