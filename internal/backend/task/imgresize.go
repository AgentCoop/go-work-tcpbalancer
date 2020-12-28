package task

import (
	"bytes"
	"encoding/gob"
	j "github.com/AgentCoop/go-work"
	r "github.com/AgentCoop/go-work-tcpbalancer/internal/common/imgresize"
	net "github.com/AgentCoop/go-work-tcpbalancer/internal/common/net"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
)

func resizeImage(j j.JobInterface, req *r.Request, ac *net.ActiveConn) {
	result := &r.Response{}
	buf := bytes.NewBuffer(req.ImgData)
	img, _, err := image.Decode(buf)
	j.Assert(err)

	m := resize.Resize(req.TargetWidth, req.TargetHeight, img, resize.Lanczos3)
	jpeg.Encode(buf, m, nil)

	result.ImgData = buf.Bytes()
	result.Width = req.TargetWidth
	result.Height = req.TargetHeight

	ac.GetWriteChan() <- result
}

func ResizeImageTask(j j.JobInterface) (func(), func() interface{}, func()) {
	run := func() interface{} {
		ac := j.GetValue().(*net.ActiveConn)
		for {
			select {
			case <-ac.GetOnNewConnChan():
			case frame := <-ac.GetOnDataFrameChan():
				buf := bytes.NewBuffer(frame)
				dec := gob.NewDecoder(buf)
				payload := &r.Request{}
				err := dec.Decode(payload)
				j.Assert(err)
				go resizeImage(j, payload, ac)
			}
		}
		return nil
	}
	return nil, run, nil
}
