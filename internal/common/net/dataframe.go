package net

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	n "net"
)

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
		var nn, head, tail int
		binary.Read(r, binary.BigEndian, &fl)
		framebuf = make([]byte, fl)
		//head = n1 - mwl - 8
		for head = copy(framebuf[0:fl], f.readbuf[mwl+8:]); head < len(framebuf); {
			fmt.Printf("Loop\n")
			nn, err = r.Read(f.readbuf)
			if err != nil { return nil, err, nil }
			head += copy(framebuf[head:], f.readbuf[:nn])
		}
		tail = nn - len(framebuf)
		if tail > 0 {
			fmt.Printf("Tail great")
			f.tail= make([]byte, tail)
			copy(f.tail[0:], f.readbuf[len(framebuf):])
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
