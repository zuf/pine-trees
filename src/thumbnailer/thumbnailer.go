package thumbnailer

import (
	"log"
	"runtime"
	"sync"

	"gopkg.in/h2non/bimg.v1"

	"github.com/zuf/pine-trees/src/raw"
)

type Thumbnailer struct {
	settings       Settings
	jobs           chan string
	results        chan []byte
	wg             sync.WaitGroup
	workersStarted bool
	rawProcessor   *raw.Raw
}

type ThumbnailSettings struct {
	Name   string
	Width  int
	Height int
}

type Settings struct {
	ThumbnailTypes []ThumbnailSettings
}

func NewThumbnailer(thumnnailTypes []ThumbnailSettings) *Thumbnailer {
	// TODO: sanity check for ThumbnailSettings

	t := Thumbnailer{
		workersStarted: false,
		rawProcessor:   raw.NewRawProcessor(),

		settings: Settings{
			ThumbnailTypes: thumnnailTypes,
		},
	}

	//t.StartWorkers()

	return &t
}

func (t *Thumbnailer) init() {
	// TODO: remove magic numbers
	jobsChanSize := 4 * runtime.NumCPU()
	resultsCacheSize := 2 * runtime.NumCPU()

	t.jobs = make(chan string, jobsChanSize)
	t.results = make(chan []byte, resultsCacheSize)
}

func (t *Thumbnailer) StartWorkers() {
	if t.workersStarted == true {
		panic("Thumbnailer workers already started!")
	}

	t.init()

	// TODO: remove magic numbers
	workersCount := 2 * runtime.NumCPU()
	t.workersStarted = true

	for w := 0; w < workersCount; w++ {
		t.wg.Add(1)
		go t.worker(t.jobs, t.results)
	}
}

func (t *Thumbnailer) Close() {
	close(t.jobs)
	t.wg.Wait()
	close(t.results)
	t.rawProcessor.Close()
}

func (t *Thumbnailer) Push(filePath string) {
	t.jobs <- filePath
}

//func (t *Thumbnailer) ResultCallback(f func(imageBuffer []byte)) {
//	var processedResults <-chan []byte
//	processedResults = t.results
//
//	for buffer := range processedResults {
//		f(buffer)
//	}
//}

func (t *Thumbnailer) ResultChan() <-chan []byte {
	return t.results
}

func (t *Thumbnailer) ProcessOne(filePath string) []byte {
	buf := t.rawProcessor.ExtractPreview(filePath, process)
	return buf
}

func (t *Thumbnailer) worker(jobs <-chan string, results chan<- []byte) {
	rawProcessor := raw.NewRawProcessor()
	defer rawProcessor.Close()

	for filePath := range jobs {
		results <- rawProcessor.ExtractPreview(filePath, process)
	}

	t.wg.Done()
}

func process(decodedImageBuffer []byte, flip int) ([]byte, error) {
	previewBuffer, err := bimg.NewImage(decodedImageBuffer).SmartCrop(300, 200)
	if err != nil {
		log.Printf("ERROR: %s", err)
	}
	return previewBuffer, err
}
