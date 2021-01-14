package frontend

import (
	"fmt"
	job "github.com/AgentCoop/go-work"
	"github.com/AgentCoop/go-work-tcpbalancer/internal/imgresize"
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
	scanneridx int
	dryRun bool
	sentx int // number of files sent for resizing
	recvx int // number of resized files received
	scandone bool
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
	// Do some initialization here
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
		case finishedTask := <- j.TaskDoneNotify(): // Wait for the scanner task to be done
			if finishedTask.GetIndex() == s.scanneridx {
				s.scandone = true
			}
			task.Tick()
		case frame := <-stream.RecvDataFrame(): // Process response from the backend server
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

			j.Log(1) <- fmt.Sprintf("file %s has been saved", filename)
			stream.RecvDataFrameSync() // Tell netmanager.Read that we are done processing the frame
			s.recvx++
			task.Tick()
		default:
			switch {
			case s.scandone && s.recvx == s.sentx: // Check if all found images were processed
				task.FinishJob()
			default:
				task.Idle() // Do nothing
			}
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
		filepath.Walk(s.inputDir, func(path string, info os.FileInfo, err error) error {
			task.Assert(err)
			stream := j.GetValue().(netmanager.Stream)

			req := &imgresize.Request{}
			req.TargetWidth = s.w
			req.TargetHeight = s.h
			req.DryRun = s.dryRun
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

			s.sentx++
			stream.Write() <- req
			stream.WriteSync()

			j.Log(1) <- fmt.Sprintf("image file %s dispatched for resizing", path)
			return nil
		})
		task.Done()
	}
	return init, run, nil
}
