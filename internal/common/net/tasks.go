package net

import (
	job "github.com/AgentCoop/go-work"
	"net"
	"sync/atomic"
)

func ListenTask(j job.Job) (func(), func() interface{}, func()) {
	init := func() {
		cm := j.GetValue().(*connManager)
		lis, err := net.Listen(cm.network, cm.addr)
		cm.plis = lis
		j.Assert(err)
	}

	run := func() interface{} {
		cm := j.GetValue().(*connManager)
		pconn, acceptErr := cm.plis.Accept()
		j.Assert(acceptErr)
		atomic.AddInt32(&cm.inboundCounter, 1)
		//pconn.SetDeadline(time.Now().Add(6 * time.Second))
		go func() {
			ac := &ActiveConn{conn: pconn, typ: Inbound, state: Active}
			ac.writeChan = make(chan interface{}, 1)
			ac.readChan = make(chan interface{}, 1)
			ac.connManager = cm
			cm.addConn(ac)
			inJob := job.NewJob(ac)
			inJob.AddTask(connWriteTask)
			inJob.AddTask(connReadTask)
			<-inJob.Run()
		}()
		return nil
	}

	cancel := func() { }
	return init, run, cancel
}

func connWriteTask(j job.Job) (func(), func() interface{}, func()) {
	run := func() interface{} {
		ac := j.GetValue().(*ActiveConn)
		var n int
		var err error
		select {
		case data := <- ac.writeChan:
			switch data.(type) {
			case []byte: // raw data
				n, err = ac.conn.Write(data.([]byte))
				j.Assert(err)
			case nil:
				a := 1
				a++
				// Handle error
			default:
				df := NewDataFrame()
				enc, err := df.Encode(data)
				j.Assert(err)
				n, err = ac.conn.Write(enc)
				j.Assert(err)
			}
			atomic.AddUint64(&ac.connManager.bytesSent, uint64(n))
			j.Assert(err)
		}
		return nil
	}
	return nil, run, nil
}

func connReadTask(j job.Job) (func(), func() interface{}, func()) {
	var readbuf []byte
	df := NewDataFrame()
	init := func() {
		readbuf = make([]byte, 1024)
	}
	run := func() interface{} {
		ac := j.GetValue().(*ActiveConn)
		frame, err, raw := df.Decode(ac.conn)
		j.Assert(err)
		evt := &Event{}
		evt.conn = ac
		if frame != nil {
			evt.data = frame
			ac.connManager.dataFrame <- evt
		} else if raw != nil {
			evt.data = raw
			ac.connManager.rawstream <- evt
		}
		return nil
	}
	cancel := func() {
		ac := j.GetValue().(*ActiveConn)
		cm := ac.connManager
		atomic.AddInt32(&cm.outboundCounter, -1)
		if ac.netJob != nil {
			ac.netJob.Cancel()
		}
		cm.delConn(ac)
	}
	return init, run, cancel
}
