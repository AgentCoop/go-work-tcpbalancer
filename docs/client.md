### Client

The main job for the client is the following:
```go
    imgResizer := frontend.NewImageResizer(CliOptions.ImgDir, CliOptions.OutputDir,
        CliOptions.Width, CliOptions.Height, CliOptions.DryRun)
    j := job.NewJob(nil)
    j.AddOneshotTask(mngr.ConnectTask)
    j.AddTask(netmanager.ReadTask)
    j.AddTask(netmanager.WriteTask)
    j.AddTask(imgResizer.ScanForImagesTask)
    j.AddTaskWithIdleTimeout(imgResizer.SaveResizedImageTask, time.Second * 8)
```
The purpose of the oneshot task _mngr.ConnectTask_ is to establish a network connection to the proxy server. Until it 
succeeds in doing that, no other tasks will be run - they all wait for _mngr.ConnectTask_ to be successfully executed.

Once it does that, it sets Job value to the network connection that will be used by two special tasks _netmanager.ReadTask_,
_netmanager.WriteTask_ that read and write network data.

Let's take a look at their implementation to get the idea of how Job tasks perform its core function:
```go
// Serves as an oneshot task to establish a TCP connection to the target host
func (mngr *connManager) ConnectTask(j job.Job) (job.Init, job.Run, job.Finalize) {
	run := func(task job.Task) {
		conn, err := net.Dial(mngr.network, mngr.addr)
		task.Assert(err)

		stream := mngr.NewStreamConn(conn, Outbound)
		j.SetValue(stream)
		mngr.addConn(stream)

		// Notify other tasks that connection was successfully established
		stream.newConnChan <- job.NotifySig
		// Finish task execution and kick off execution of other tasks.
		task.Done()
	}
	return nil, run, nil
}
```

```go
func read(stream *stream, task job.Task) {
	n, err := stream.conn.Read(stream.readbuf) // Read network data
	task.Assert(err) // Assert that there is no error. A failed assertion will stop job execution.

	atomic.AddUint64(&stream.connManager.perfmetrics.bytesReceived, uint64(n))
	data := stream.readbuf[0:n]
	// Send read data in raw stream or in data frames using special task channels
	switch stream.dataKind {
	case DataFrameKind:
		frames, err := stream.framerecv.Capture(data)
		task.Assert(err)
		for _, frame := range frames {
			// Ping/Pong synchronization. Send a frame to the channel and wait for a notification from another task
			// that it was processed.
			stream.recvDataFrameChan <- frame
			<-stream.recvDataFrameSyncChan
		}
	case DataRawKind:
		stream.recvRawChan <- data
		<-stream.recvRawSyncChan
	}
	// Tick and wait for new data
	task.Tick()
}

func ReadTask(j job.Job) (job.Init, job.Run, job.Finalize) {
	run := func(task job.Task) {
		s := j.GetValue().(*stream)
		read(s, task)
	}
	fin := func(task job.Task) {
		s := j.GetValue()
		if s == nil { return }
		readFin(s.(*stream), task)
	}
	return nil, run, fin
}
```
Job tasks share data and orchestrate their execution using channels and the aforementioned ping/pong synchronization
technique. You will see soon why it's very important to use such kind of synchronization to avoid possible pitfalls.

The same sync technique is being used for _netmanager.WriteTask_:
```go
func write(s *stream, task job.Task) {
	var n int
	var err error

	select {
	case data := <- s.writeChan: // Some task "asked" us to send network data
		task.AssertNotNil(data)
		switch data.(type) {
		case []byte: // raw data
			n, err = s.conn.Write(data.([]byte))
			task.Assert(err)
		default:
			enc, err := netdataframe.ToFrame(data)
			task.Assert(err)
			n, err = s.conn.Write(enc.GetBytes())
			task.Assert(err)
		}
		// Sync with the writer
		s.writeSyncChan <- n // Tell that task that we are done sending data
		atomic.AddUint64(&s.connManager.perfmetrics.bytesSent, uint64(n))
		task.Tick()
	}
}

func WriteTask(j job.Job) (job.Init, job.Run, job.Finalize) {
	run := func(task job.Task) {
		s := j.GetValue().(*stream)
		write(s, task)
	}
	fin := func(task job.Task)  {
		s := j.GetValue()
		if s == nil { return }
		writeFin(s.(*stream), task)
	}
	return nil, run, fin
}
```
Finally, the other two tasks ([imgResizer.ScanForImagesTask](https://github.com/AgentCoop/go-work-tcpbalancer/blob/main/internal/task/frontend/imgresize.go#L87)
and [mgResizer.SaveResizedImageTask](https://github.com/AgentCoop/go-work-tcpbalancer/blob/main/internal/task/frontend/imgresize.go#L38))
do the rest of things: the former dispatches an image to the proxy server, and the latter saves the resized image. They
both share data and orcherstrate execution with each other.
