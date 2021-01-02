package frontend

import (
	"bytes"
	"encoding/gob"
	"fmt"
	job "github.com/AgentCoop/go-work"
	"github.com/AgentCoop/go-work-tcpbalancer/internal/common/imgresize"
	"github.com/AgentCoop/go-work-tcpbalancer/internal/common/net"
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
	foundCounter int
	savedCounter int
	done bool
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
func (s *ImageResizer) SaveResizedImageTask(j job.JobInterface) (job.Init, job.Run, job.Cancel) {
	init := func(t *job.TaskInfo) {
		if _, err := os.Stat(s.inputDir); os.IsNotExist(err) {
			j.Assert(err)
		}
		if _, err := os.Stat(s.outputDir); os.IsNotExist(err) {
			err := os.Mkdir(s.outputDir, 755)
			j.Assert(err)
		}
	}
	run := func(t *job.TaskInfo) {
		ac := j.GetValue().(*net.ActiveConn)
		select {
		case dataFrame := <-ac.GetOnDataFrameChan():
			res := &imgresize.Response{}
			buf := bytes.NewBuffer(dataFrame)
			dec := gob.NewDecoder(buf)
			err := dec.Decode(res)
			j.Assert(err)

			baseName := fmt.Sprintf("%s-%dx%d%s",
				res.OriginalName, res.ResizedWidth, res.ResizedHeight, res.Typ.ToFileExt())
			filename := s.outputDir + string(os.PathSeparator) + baseName
			ioutil.WriteFile(filename, res.ImgData, 0775)
			s.savedCounter++

			ac.OnDataFrameDoneChan <- struct{}{}
			j.Log(1) <- fmt.Sprintf("[ save-task ]: file %s has been saved\n", filename)
		default:
			// Finish job
			if s.done && s.savedCounter >= s.foundCounter {
				t.Done()
				j.Finish()
			}
		}
		t.Tick()
	}
	return init, run, nil
}

// Scans the given directory for images to resize.
func (s *ImageResizer) ScanForImagesTask(j job.JobInterface) (job.Init, job.Run, job.Cancel) {
	init := func(t *job.TaskInfo) {
		t.SetResult(0) // scanned images counter
	}
	run := func(t *job.TaskInfo) {
		req := &imgresize.Request{}
		req.TargetWidth = s.w
		req.TargetHeight = s.h
		filepath.Walk(s.inputDir, func(path string, info os.FileInfo, err error) error {
			j.Assert(err)
			ac := j.GetValue().(*net.ActiveConn)

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
			j.Assert(err)
			req.ImgData = data

			ac.GetWriteChan() <- req
			<-ac.GetWriteDoneChan()

			s.foundCounter++
			j.Log(1) <- fmt.Sprintf("[ scanner-task ]: image file %s dispatched for resizing\n", path)
			return nil
		})
		s.done = true
		t.Done()
	}
	return init, run, func() { }
}
