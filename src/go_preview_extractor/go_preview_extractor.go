package go_preview_extractor

import (
	"context"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/xor-gate/goexif2/exif"
)

func ShotTime(fname string) (time.Time, error) {
	var tm time.Time

	f, err := os.Open(fname)
	defer f.Close() // TODO: should verify errors on file close

	if err != nil {
		log.Println(err)
		return tm, err
	}

	// Optionally register camera makenote data parsing - currently Nikon and
	// Canon are supported.
	//exif.RegisterParsers(mknote.All...)

	x, err := exif.Decode(f)
	if err != nil {
		log.Println(err)
		return tm, err
	}

	//camModel, _ := x.Get(exif.Model) // normally, don't ignore errors!
	//fmt.Println(camModel.StringVal())
	//

	//focal, _ := x.Get(exif.FocalLength)
	//numer, denom, _ := focal.Rat2(0) // retrieve first (only) rat. value
	//fmt.Printf("%v/%v", numer, denom)
	//
	//Two convenience functions exist for date/time taken and GPS coords:
	// TODO: look for timezone offset, GPS time, etc.
	tm, err = x.DateTime()
	if err != nil {
		log.Println(err)
		return tm, err
	}

	return tm, nil
	//fmt.Println("Taken: ", tm)

	//lat, long, _ := x.LatLong()
	//fmt.Println("lat, long: ", lat, ", ", long)

}

func JPEGPreviewFromExif(fname string, fullPreview bool) ([]byte, error) {

	f, err := os.Open(fname)
	defer f.Close() // TODO: should verify errors on file close

	if err != nil {
		log.Println(err)
		return nil, err
	}

	// Optionally register camera makenote data parsing - currently Nikon and
	// Canon are supported.
	//exif.RegisterParsers(mknote.All...)

	x, err := exif.Decode(f)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	//camModel, _ := x.Get(exif.Model) // normally, don't ignore errors!
	//fmt.Println(camModel.StringVal())
	//
	//focal, _ := x.Get(exif.FocalLength)
	//numer, denom, _ := focal.Rat2(0) // retrieve first (only) rat. value
	//fmt.Printf("%v/%v", numer, denom)
	//
	//Two convenience functions exist for date/time taken and GPS coords:
	//tm, _ := x.DateTime()
	//fmt.Println("Taken: ", tm)

	//lat, long, _ := x.LatLong()
	//fmt.Println("lat, long: ", lat, ", ", long)

	var bytesBuf []byte

	if fullPreview {
		bytesBuf, err = x.PreviewImage()

		if err != nil {
			// maybe it jpeg?
			log.Println(err)

			bytesBuf, err = ioutil.ReadFile(fname)
			if err != nil {
				log.Println(err)
				return nil, err
			}

			return bytesBuf, nil
		}
	} else {
		bytesBuf, err = x.JpegThumbnail()
	}
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return bytesBuf, nil

}

type workerTask struct {
	file        string
	fullPreview bool

	ctx context.Context

	writer *io.PipeWriter
	reader *io.PipeReader
}

func (w *workerTask) GetReader() io.Reader {
	return w.reader
}

func NewWorkerTask(ctx context.Context, file string, fullPreview bool) *workerTask {
	r, w := io.Pipe()

	return &workerTask{
		file:        file,
		writer:      w,
		reader:      r,
		fullPreview: fullPreview,
		ctx:         ctx,
	}
}

type WorkerPool struct {
	inputs chan *workerTask
	wg     *sync.WaitGroup
}

func NewWorkersPool(n int) *WorkerPool {
	jobsChanSize := n * 2

	if jobsChanSize < 2 {
		jobsChanSize = 2
	}

	inputChan := make(chan *workerTask, jobsChanSize)
	wg := sync.WaitGroup{}

	workersPool := WorkerPool{
		inputs: inputChan,
		wg:     &wg,
	}

	workersCount := n

	if n <= 0 {
		workersCount = runtime.NumCPU()
	}

	for w := 0; w < workersCount; w++ {
		wg.Add(1)
		go worker(inputChan, &wg)
	}

	return &workersPool
}

func (w *WorkerPool) Close() {
	close(w.inputs)
	w.wg.Wait()
}

func (w *WorkerPool) ProcessFile(ctx context.Context, file string, fullPreview bool) io.Reader {
	fileToProcess := file

	if strings.ToLower(path.Ext(file)) == ".xmp" {
		fileToProcess = file[:len(file)-4]
	}

	task := NewWorkerTask(ctx, fileToProcess, fullPreview)
	w.inputs <- task

	return task.GetReader()
}

func worker(inputs <-chan *workerTask, wg *sync.WaitGroup) {
	defer wg.Done()

	for task := range inputs {
		//s := fmt.Sprintf("%s - done!", task.file)
		select {
		case <-task.ctx.Done():
			log.Printf("request cancelled because: %+v\n", task.ctx.Err())

			err := task.writer.Close()
			if err != nil {
				log.Printf("ERROR: can't close writer after processing file \"%s\": %s", task.file, err)
			}
			continue
		default:

			buf, err := JPEGPreviewFromExif(task.file, task.fullPreview)
			if err != nil {
				log.Printf("ERROR: can't extract thumbnail from \"%s\": %s", task.file, err)

				// TODO: DRY
				err = task.writer.Close()
				if err != nil {
					log.Printf("ERROR: can't close writer after processing file \"%s\": %s", task.file, err)
				}
				continue
			}

			_, err = task.writer.Write(buf)
			if err != nil {
				log.Printf("ERROR: can't write to writer after processing file \"%s\": %s", task.file, err)
			}

			err = task.writer.Close()
			if err != nil {
				log.Printf("ERROR: can't close writer after processing file \"%s\": %s", task.file, err)
			}
		}
	}
}
