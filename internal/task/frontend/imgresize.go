package frontend

import (
	"fmt"
	job "github.com/AgentCoop/go-work"
	"github.com/AgentCoop/go-work-tcpbalancer/internal/common/imgresize"
	"github.com/AgentCoop/net-manager"
	"io/ioutil"
	"mime"
	"time"

	//"math/rand"
	"os"
	"path/filepath"
)

type ImageResizer struct {
	inputDir string
	outputDir string
	w uint
	h uint
	scanneridx int
	dryRun bool
	filescount int
}

func NewImageResizer(input string, output string, w uint, h uint, dryRun bool) *ImageResizer {
	s := &ImageResizer{
		inputDir:  input,
		outputDir: output,
		w: w,
		h: h,
		dryRun: dryRun,
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
		case frame := <-stream.RecvDataFrame():
			task.AssertNotNil(frame)
			res := &imgresize.Response{}
			err := frame.Decode(res)
			task.Assert(err)

			baseName := fmt.Sprintf("%s-%dx%d%s",
				res.OriginalName, res.ResizedWidth, res.ResizedHeight, res.Typ.ToFileExt())
			filename := s.outputDir + string(os.PathSeparator) + baseName
			if ! s.dryRun {
				ioutil.WriteFile(filename, res.ImgData, 0775)
			}

			j.Log(1) <- fmt.Sprintf("[ save-task ]: file %s has been saved\n", filename)
			stream.RecvDataFrameSync()

			fmt.Printf("%d %d\n", res.ImgIndex, s.filescount)

			if res.ImgIndex == s.filescount - 1 {
				j.Finish()
			} else {
				task.Tick()
			}
		//default:
		//	task.Idle()
		}
	}
	return init, run, nil
}

// Scans the given directory for images to resize.
func (s *ImageResizer) ScanForImagesTask(j job.Job) (job.Init, job.Run, job.Finalize) {
	init := func(task job.Task) {
		s.scanneridx = task.GetIndex()
	}
	run := func(task job.Task) {
		req := &imgresize.Request{}
		req.TargetWidth = s.w
		req.TargetHeight = s.h
		req.DryRun = s.dryRun
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

			if ! s.dryRun {
				data, err := ioutil.ReadFile(path)
				task.Assert(err)
				req.ImgData = data
			}

			req.ImgIndex = s.filescount
			s.filescount++

			stream.Write() <- req
			stream.WriteSync()

			time.Sleep(time.Millisecond * 10)
			//fmt.Printf("[ scanner-task ]: image file %s dispatched for resizing\n", path)
			j.Log(1) <- fmt.Sprintf("[ scanner-task ]: image file %s dispatched for resizing\n", path)
			return nil
		})
		task.Done()
	}
	return init, run, nil
}
