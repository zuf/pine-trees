package main

import (
	"bufio"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/zuf/pine-trees/raw"

	"github.com/labstack/echo/middleware"

	"github.com/labstack/echo"

	"gopkg.in/h2non/bimg.v1"

	"github.com/zuf/pine-trees/thumbnailer"
)

//func DecodeThumb(raw2 *raw.Raw, filePath string) []byte {
//	return raw2.ExtractPreview(filePath)
//}

// TODO: remove global var
var t time.Time

func printStats(count int, startedAt time.Time) {
	if time.Since(t) > 5*time.Second {
		duration := time.Since(startedAt)
		fmt.Printf("Processes %d images in %s (%.2fms avg per image, images per second: %.2f)\n",
			count,
			duration.Round(time.Millisecond),
			((float64)(duration.Nanoseconds())/((float64)(count)))/((float64)(time.Millisecond)),
			float64(count)/duration.Seconds())

		t = time.Now()
	}
}

func processFromStdin() {
	fmt.Println("Start reading image paths from stdin")

	thmb := thumbnailer.NewThumbnailer([]thumbnailer.ThumbnailSettings{
		{Width: 300, Height: 200},
	})
	defer thmb.Close()

	thmb.StartWorkers()

	t = time.Now().Add(-1 * time.Minute)
	fmt.Println()

	go func() {
		n := 0
		startedAt := time.Now()
		for imageBuffer := range thmb.ResultChan() {
			name := fmt.Sprintf("/tmp/images_test/new_test_%d.jpg", n)
			err := bimg.Write(name, imageBuffer)

			if err != nil {
				fmt.Printf("ERROR: %s", err)
				panic(err)
			}

			n++
			printStats(n, startedAt)
		}
		printStats(n, startedAt)
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		thmb.Push(line)
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}

}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

type Photo struct {
	Src             string
	Name            string
	Directory       bool
	SupportedFormat bool
	SendToBrowser   bool
}

type PhotosPageData struct {
	Photos  []Photo
	Page    int
	MaxPage int
	Pages   []int
	Path    string
	DirPath string
}

func basePath() string {
	return os.Getenv("GALLERY_PATH")
}

func fullPath(path string) string {
	dir := ""

	if len(path) > 0 {
		dir = filepath.Join(basePath(), path)
	} else {
		return basePath()
	}

	return dir
}

func Index(c echo.Context) error {
	path := c.QueryParam("s")
	data := PhotosPageData{Photos: []Photo{}}

	// TODO: use reader
	files, err := ioutil.ReadDir(fullPath(path))
	if err != nil {
		log.Fatal(err)
	}

	sort.SliceStable(files, func(i, j int) bool {
		if files[i].IsDir() {
			if files[j].IsDir() {
				return files[i].Name() < files[j].Name()
			} else {
				return true
			}
		} else {
			if files[j].IsDir() {
				return false
			} else {
				return files[i].Name() < files[j].Name()
			}
		}
	})

	page := 1
	page, _ = strconv.Atoi(c.QueryParam("p"))
	if page < 1 {
		page = 1
	}
	perPage := 60
	from := (page - 1) * perPage
	to := from + perPage
	if to > len(files) {
		to = len(files)
	}

	data.Page = page
	data.MaxPage = len(files) / perPage

	data.Pages = []int{}
	data.Path = path
	data.DirPath = filepath.Dir(path)

	for n := 1; n <= data.MaxPage; n++ {
		data.Pages = append(data.Pages, n)
	}

	for _, f := range files[from:to] {
		fullPath := filepath.Join(path, f.Name())

		if f.IsDir() {
			data.Photos = append(data.Photos, Photo{Src: fullPath, Name: f.Name(), Directory: true})
		} else {
			if f.Mode().IsRegular() {
				ext := strings.ToUpper(filepath.Ext(f.Name()))
				if ext == ".CR2" {
					data.Photos = append(data.Photos, Photo{Src: fullPath, Name: f.Name(), SupportedFormat: true, SendToBrowser: false})
				} else {
					data.Photos = append(data.Photos, Photo{Src: fullPath, Name: f.Name(), SupportedFormat: false, SendToBrowser: true})
				}
			}
		}
	}

	return c.Render(http.StatusOK, "photos", data)
}

func process(decodedImageBuffer []byte, flip int) ([]byte, error) {
	//if flip == 0 {
	previewBuffer, err := bimg.NewImage(decodedImageBuffer).SmartCrop(300, 200)
	if err != nil {
		log.Printf("ERROR: %s", err)
	}
	return previewBuffer, err
	//} else {
	//
	//	image := bimg.NewImage(decodedImageBuffer)
	//	options := bimg.Options{
	//		Width:  300,
	//		Height: 200,
	//		Crop:   true,
	//
	//		Gravity: bimg.GravitySmart,
	//	}
	//
	//	switch flip {
	//	case 1:
	//	case 2:
	//		options.Flip = true
	//
	//	case 3:
	//		options.Rotate = 180
	//	case 4:
	//		options.Rotate = 180
	//		options.Flip = true
	//	case 5:
	//		options.Rotate = 270
	//		options.Flip = true
	//	case 6:
	//		options.Rotate = 270
	//	case 7:
	//		options.Rotate = 90
	//		options.Flip = true
	//	case 8:
	//		options.Rotate = 90
	//	}
	//
	//	return image.Process(options)
	//
	//}
}

func PreviewPhotoHandler(c echo.Context) error {
	filePath := fullPath(c.QueryParam("s"))
	rawProcessor := raw.NewRawProcessor()
	defer rawProcessor.Close()
	data := rawProcessor.ExtractPreview(filePath, process)

	return c.Blob(http.StatusOK, "image/jpeg", data)
}

func processFull(decodedImageBuffer []byte, flip int) ([]byte, error) {
	// TODO: rotate photo if needed
	tmp := make([]byte, len(decodedImageBuffer))
	copy(tmp, decodedImageBuffer)
	return tmp, nil
	//
	//buf, err := bimg.NewImage(decodedImageBuffer).SmartCrop(1920, 1200)
	//if err != nil {
	//	log.Printf("ERROR: %s", err)
	//}
	//return buf, err
}

func FetchHandler(c echo.Context) error {
	filePath := fullPath(c.QueryParam("s"))

	return c.File(filePath)
}

func FullPreviewPhotoHandler(c echo.Context) error {
	filePath := fullPath(c.QueryParam("s"))
	rawProcessor := raw.NewRawProcessor()
	defer rawProcessor.Close()
	data := rawProcessor.ExtractPreview(filePath, processFull)

	return c.Blob(http.StatusOK, "image/jpeg", data)
}

func main() {

	if len(basePath()) < 1 {
		fmt.Fprintf(os.Stderr, "Please set path to gallery in GALLERY_PATH env var!")
		os.Exit(1)
	}

	tpl := &Template{
		templates: template.Must(template.ParseGlob("public/views/*.html")),
	}

	e := echo.New()
	e.Static("/css", "static/css")
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Renderer = tpl
	e.GET("/", Index)
	e.GET("/d", Index)
	e.GET("/p", PreviewPhotoHandler)
	e.GET("/g", FullPreviewPhotoHandler)
	e.GET("/f", FetchHandler)
	e.Logger.Fatal(e.Start(":1323"))

}
