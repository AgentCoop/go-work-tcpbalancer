package frontend

import (
	"fmt"
	job "github.com/AgentCoop/go-work"
	"github.com/AgentCoop/go-work-tcpbalancer/internal/common/imgresize"
	"github.com/AgentCoop/net-manager"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
)

type ImageResizer struct {
	inputDir string
	outputDir string
	w uint
	h uint
}

func NewImageResizer(input string, output string, w uint, h uint) *ImageResizer {
	s := &ImageResizer{
		inputDir:  input,
		outputDir: output,
		w: w,
		h: h,
	}
	return s
}

// Saves resized image to the output dir
func (s *ImageResizer) SaveResizedImageTask(j job.Job) (job.Init, job.Run, job.Finalize) {
	init := func(t job.Task) {
		if _, err := os.Stat(s.inputDir); os.IsNotExist(err) {
			t.Assert(err)
		}
		if _, err := os.Stat(s.outputDir); os.IsNotExist(err) {
			err := os.Mkdir(s.outputDir, 755)
			t.Assert(err)
		}
	}
	run := func(task job.Task) {
		stream := j.GetValue().(netmanager.Stream)
		select {
		case dataFrame := <-stream.RecvDataFrame():
			if dataFrame == nil {
				fmt.Printf(" -->>>  nil data frame\n")
			}
			res := &imgresize.Response{}
			err := dataFrame.Decode(res)
			task.Assert(err)

			baseName := fmt.Sprintf("%s-%dx%d%s",
				res.OriginalName, res.ResizedWidth, res.ResizedHeight, res.Typ.ToFileExt())
			filename := s.outputDir + string(os.PathSeparator) + baseName
			ioutil.WriteFile(filename, res.ImgData, 0775)
			stream.RecvDataFrameSync()

			j.Log(1) <- fmt.Sprintf("[ save-task ]: file %s has been saved\n", filename)
			task.Tick()
		default:
			task.Idle()
		}
	}
	return init, run, nil
}

// Scans the given directory for images to resize.
func (s *ImageResizer) ScanForImagesTask(j job.Job) (job.Init, job.Run, job.Finalize) {
	init := func(task job.Task) {
		task.SetResult(0) // scanned images counter
	}
	run := func(task job.Task) {
		req := &imgresize.Request{}
		req.TargetWidth = s.w
		req.TargetHeight = s.h
		filepath.Walk(s.inputDir, func(path string, info os.FileInfo, err error) error {
			task.Assert(err)
			stream := j.GetValue().(netmanager.Stream)

			req.OriginalName = info.Name()
			fileExt := filepath.Ext(info.Name())
			switch mime.TypeByExtension(fileExt) {
			case "image/jpeg":
				req.Typ = imgresize.Jpeg
			case "image/png":
				req.Typ = imgresize.Png
			default:
				return nil
			}

			data, err := ioutil.ReadFile(path)
			task.Assert(err)
			req.ImgData = data

			stream.Write() <- req
			stream.WriteSync()

			j.Log(1) <- fmt.Sprintf("[ scanner-task ]: image file %s dispatched for resizing\n", path)
			return nil
		})
		task.Done()
	}
	return init, run, nil
}
