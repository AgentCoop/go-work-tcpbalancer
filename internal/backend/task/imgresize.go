package task

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/AgentCoop/go-work"
	r "github.com/AgentCoop/go-work-tcpbalancer/internal/common/imgresize"
	net "github.com/AgentCoop/go-work-tcpbalancer/internal/common/net"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"image/png"
)

func resizeImage(t *job.TaskInfo, req *r.Request, ac *net.ActiveConn) {
	result := &r.Response{}
	buf := bytes.NewBuffer(req.ImgData)
	img, _, err := image.Decode(buf)
	t.Assert(err)

	m := resize.Resize(req.TargetWidth, req.TargetHeight, img, resize.Lanczos3)
	switch req.Typ {
	case r.Jpeg:
		jpeg.Encode(buf, m, nil)
	case r.Png:
		png.Encode(buf, m)
	}

	result.ImgData = buf.Bytes()
	result.Typ = req.Typ
	result.OriginalName = req.OriginalName
	result.ResizedWidth = req.TargetWidth
	result.ResizedHeight = req.TargetHeight

	ac.GetWriteChan() <- result
	<-ac.GetWriteDoneChan()
}

func ResizeImageTask(j job.JobInterface) (job.Init, job.Run, job.Cancel) {
	run := func(t *job.TaskInfo) {
		ac := j.GetValue().(*net.ActiveConn)
		select {
		case <-ac.GetOnNewConnChan():
		case frame := <-ac.GetOnDataFrameChan():
			fmt.Printf("new frame %d bytes\n", len(frame))
			buf := bytes.NewBuffer(frame)
			dec := gob.NewDecoder(buf)
			payload := &r.Request{}
			err := dec.Decode(payload)
			t.Assert(err)
			resizeImage(t, payload, ac)
			ac.OnDataFrameDoneChan <- struct{}{}
		default:
		}
		t.Tick()
	}
	return nil, run, func() { }
}
