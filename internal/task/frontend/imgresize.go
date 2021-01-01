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
		//res := &imgresize.Response{}
		//fmt.Printf("run resize\n")
		ac := j.GetValue().(*net.ActiveConn)
		select {
		//case  <- t.GetDepChan():
			//fmt.Printf("Got dep %v\n", dep)
		case dataFrame := <-ac.GetOnDataFrameChan():
			fmt.Printf("new frame %d saved %d found %d\n", len(dataFrame), s.savedCounter, s.foundCounter)
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

			// Finish job
		default:
			if s.done && s.savedCounter >= s.foundCounter {
				fmt.Printf("Finish job\n")
				t.Done()
				j.Finish()
			}
				//fmt.Printf("nope\n")
		}
		t.Tick()
	}
	return init, run, func() {
		fmt.Println("cancel save")
	}
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

			fmt.Printf("file %s %v\n", path, err)
			data, err := ioutil.ReadFile(path)
			j.Assert(err)
			req.ImgData = data

			ac.GetWriteChan() <- req
			<-ac.GetWriteDoneChan()

			//time.Sleep(time.Millisecond * 450)
			s.foundCounter++
			return nil
		})
		s.done = true
		t.Done()
	}
	return init, run, func() { }
}
