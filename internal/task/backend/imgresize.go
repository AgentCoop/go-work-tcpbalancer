package backend

import (
	"bytes"
	"fmt"
	"github.com/AgentCoop/go-work"
	r "github.com/AgentCoop/go-work-tcpbalancer/internal/common/imgresize"
	"github.com/AgentCoop/net-manager"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"image/png"
)

func resizeImage(t job.Task, req *r.Request, stream netmanager.Stream) {
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

	stream.Write() <- result
	stream.WriteSync()
}

func ResizeImageTask(j job.Job) (job.Init, job.Run, job.Finalize) {
	run := func(task job.Task) {
		stream := j.GetValue().(netmanager.Stream)
		select {
		case frame := <-stream.RecvDataFrame():
			payload := &r.Request{}
			err := frame.Decode(payload)
			task.Assert(err)

			resizeImage(task, payload, stream)
			stream.RecvDataFrameSync()
		//default:
		}
		//j.Finish()
		task.Tick()
	}
	return nil, run, func(task job.Task) {
		_, err := task.GetJob().GetInterruptedBy()
		fmt.Printf("close by %s\n", err)
	}
}
