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
	"time"
)

type ResizerOptions struct { }

func resizeImage(t job.Task, imgData []byte, typ r.ImgType, w uint, h uint) []byte {
	buf := bytes.NewBuffer(imgData)
	img, _, err := image.Decode(buf)
	t.Assert(err)

	m := resize.Resize(w, h, img, resize.Lanczos3)
	switch typ {
	case r.Jpeg:
		jpeg.Encode(buf, m, nil)
	case r.Png:
		png.Encode(buf, m)
	}

	return buf.Bytes()
}

func (o *ResizerOptions) ResizeImageTask(j job.Job) (job.Init, job.Run, job.Finalize) {
	run := func(task job.Task) {
		stream := j.GetValue().(netmanager.Stream)
		select {
		case frame := <-stream.RecvDataFrame():
			task.AssertNotNil(frame)
			req := &r.Request{}
			resp := &r.Response{}
			resp.CreatedAt = time.Now().UnixNano()
			err := frame.Decode(req)
			task.Assert(err)

			resp.Typ = req.Typ
			resp.ResizedWidth = req.TargetWidth
			resp.ResizedHeight = req.TargetHeight
			resp.OriginalName = req.OriginalName

			if ! req.DryRun {
				start := time.Now().UnixNano()
				resp.ImgData = resizeImage(task, req.ImgData, req.Typ, req.TargetWidth, req.TargetHeight)
				end := time.Now().UnixNano()
				resp.ProcessingTime = end - start
			}

			j.Log(1) <- fmt.Sprintf("image %s has been resized", resp.OriginalName)

			stream.Write() <- resp
			stream.WriteSync()
			stream.RecvDataFrameSync()
		//default:
		//	task.Idle()
		//	return
		}
		task.Tick()
	}
	return nil, run, nil
}
