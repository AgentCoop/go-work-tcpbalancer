package task

import (
	"bytes"
	"encoding/gob"
	"fmt"
	job "github.com/AgentCoop/go-work"
	"github.com/AgentCoop/go-work-tcpbalancer/internal/common/imgresize"
	"github.com/AgentCoop/go-work-tcpbalancer/internal/common/net"
	"github.com/AgentCoop/go-work-tcpbalancer/internal/frontend"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
)

// Saves resized image to the output dir
func SaveResizedImageTask(j job.JobInterface) (job.Init, job.Run, job.Cancel) {
	init := func() {
		if _, err := os.Stat(frontend.ImgResizeOptions.ImgDir); os.IsNotExist(err) {
			j.Assert(err)
		}
		if _, err := os.Stat(frontend.ImgResizeOptions.OutputDir); os.IsNotExist(err) {
			err := os.Mkdir(frontend.ImgResizeOptions.OutputDir, 755)
			j.Assert(err)
		}
	}
	run := func(t *job.TaskInfo) interface{} {
		res := &imgresize.Response{}
		ac := j.GetValue().(*net.ActiveConn)
		select {
		case dataFrame := <-ac.GetOnDataFrameChan():
			buf := bytes.NewBuffer(dataFrame)
			dec := gob.NewDecoder(buf)
			err := dec.Decode(res)
			j.Assert(err)

			baseName := fmt.Sprintf("%s-%dx%d.%s",
				res.OriginalName, res.ResizedWidth, res.ResizedHeight, res.Typ.ToFileExt())
			filename := frontend.ImgResizeOptions.OutputDir + string(os.PathSeparator) + baseName
			ioutil.WriteFile(filename, res.ImgData, 775)
		}
		return nil
	}
	return init, run, func() {
		fmt.Println("cancel save")
	}
}

type ImageScanner struct {
	inputDir string
	outputDir string
}

func NewImageScanner(input string, output string) *ImageScanner {
	s := &ImageScanner{
		inputDir:  input,
		outputDir: output,
	}
	return s
}

// Scans the given directory for images to resize.
func ScanForImagesTask(j job.JobInterface) (job.Init, job.Run, job.Cancel) {
	run := func(t *job.TaskInfo) interface{} {
		req := &imgresize.Request{}
		req.TargetWidth = frontend.ImgResizeOptions.Width
		req.TargetHeight = frontend.ImgResizeOptions.Height
		filepath.Walk(frontend.ImgResizeOptions.ImgDir, func(path string, info os.FileInfo, err error) error {
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
			fmt.Printf("wait for write done\n")
			//<-ac.GetWriteDoneChan()
			//fmt.Printf("wdone\n")
			//time.Sleep(time.Second)
			return nil
		})
		return true
	}
	return nil, run, func() {
		fmt.Println("cancel scanner")
	}
}
