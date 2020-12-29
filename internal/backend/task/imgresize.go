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

func resizeImage(j job.JobInterface, req *r.Request, ac *net.ActiveConn) {
	result := &r.Response{}
	buf := bytes.NewBuffer(req.ImgData)
	img, _, err := image.Decode(buf)
	j.Assert(err)

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

	fmt.Printf("write to chan\n")
	ac.GetWriteChan() <- result
	//n := <-ac.GetWriteDoneChan()
	//fmt.Printf("done write %d\n", n)
}

func ResizeImageTask(j job.JobInterface) (job.Init, job.Run, job.Cancel) {
	run := func(t *job.TaskInfo) interface{} {
		ac := j.GetValue().(*net.ActiveConn)
		for {
			fmt.Printf("wait for frame\n")
			select {
			case <-ac.GetOnNewConnChan():
			case frame := <-ac.GetOnDataFrameChan():
				fmt.Printf("new frame\n")
				buf := bytes.NewBuffer(frame)
				dec := gob.NewDecoder(buf)
				payload := &r.Request{}
				err := dec.Decode(payload)
				j.Assert(err)
				resizeImage(j, payload, ac)
			}
		}
		return nil
	}
	return nil, run, func() {
		fmt.Printf("Cancel resize job\n")
	}
}
