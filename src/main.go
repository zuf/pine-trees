package main

import (
	"bufio"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	bimg2 "gopkg.in/h2non/bimg.v1"

	"github.com/karlseguin/ccache"

	"github.com/zuf/pine-trees/src/raw"

	"github.com/labstack/echo/middleware"

	"github.com/labstack/echo"

	"github.com/zuf/pine-trees/src/thumbnailer"
)

//func DecodeThumb(raw2 *raw.Raw, filePath string) []byte {
//	return raw2.ExtractPreview(filePath)
//}

// TODO: remove global var
var t time.Time
var thumbnailsCache *ccache.Cache
var filesListCache *ccache.Cache

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
			err := bimg2.Write(name, imageBuffer)

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
	IsVideo         bool
	VideoPreview    string
}

type PageNumber struct {
	Num     int
	Current bool
}

type BreadCrumb struct {
	Title string
	Path  string
}

type PhotosPageData struct {
	Photos      []Photo
	Page        int
	MaxPage     int
	Pages       []PageNumber
	Path        string
	DirPath     string
	BreadCrumbs []BreadCrumb
	PrevPage    int
	NextPage    int
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
	page := 1
	page, _ = strconv.Atoi(c.QueryParam("p"))

	h := md5.New()
	io.WriteString(h, path)
	io.WriteString(h, strconv.Itoa(page))
	dirName := fullPath(path)

	fi, err := os.Stat(dirName)
	if err != nil {
		c.Logger().Errorf("Can't read directory \"%s\": %s", dirName, err)

		//return err
		return echo.NewHTTPError(http.StatusNotFound, "Directory not found")
	}

	io.WriteString(h, strconv.FormatInt(fi.Size(), 36))
	io.WriteString(h, strconv.FormatInt(fi.ModTime().UTC().UnixNano(), 36))

	c.Response().Header().Set("Cache-Control", "max-age=60")
	// TODO: last mod time in this dir for las modified file instead of directory or min(lastFileModTime, dirModTilme)?
	c.Response().Header().Set("Last-Modified", fi.ModTime().UTC().Format(http.TimeFormat))
	//c.Response().Header().Set("Last-Modified", fi.ModTime().UTC().Format(http.TimeFormat))

	etagKey := hex.EncodeToString(h.Sum(nil))
	c.Response().Header().Set("ETag", etagKey)

	item, err := filesListCache.Fetch(etagKey, time.Second*60, func() (interface{}, error) {
		data := PhotosPageData{Photos: []Photo{}}

		// TODO: use reader
		//files, err := ioutil.ReadDir(fullPath(path))

		f, err := os.Open(dirName)
		defer f.Close()

		if err != nil {
			return nil, err
		}
		files, err := f.Readdir(-1)
		if err != nil {
			return nil, err
		}

		sort.Slice(files, func(i, j int) bool {
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

		if page < 1 {
			page = 1
		}
		perPage := 60
		from := (page - 1) * perPage
		to := from + perPage

		if from > len(files) {
			from = len(files)
		}

		if to > len(files) {
			to = len(files)
		}

		data.Page = page
		data.MaxPage = len(files) / perPage

		if len(files)%perPage != 0 {
			data.MaxPage += 1
		}

		data.Pages = []PageNumber{}
		data.Path = path
		data.DirPath = filepath.Dir(path)

		data.BreadCrumbs = []BreadCrumb{}

		data.PrevPage = data.Page - 1
		if data.PrevPage < 1 {
			data.PrevPage = 1
		}
		data.NextPage = data.Page + 1
		if data.NextPage > data.MaxPage {
			data.NextPage = data.MaxPage
		}

		prev_p := ""
		for _, p := range strings.Split(data.Path, string(os.PathSeparator)) {
			itemPath := filepath.Join(prev_p, p)
			bc := BreadCrumb{
				Title: p,
				Path:  itemPath,
			}
			data.BreadCrumbs = append(data.BreadCrumbs, bc)
			prev_p = itemPath
		}

		for n := 1; n <= data.MaxPage; n++ {
			data.Pages = append(data.Pages, PageNumber{Num: n, Current: page == n})
		}

		for _, f := range files[from:to] {
			fullPath := filepath.Join(path, f.Name())

			if f.IsDir() {
				data.Photos = append(data.Photos, Photo{Src: fullPath, Name: f.Name(), Directory: true, IsVideo: false})
			} else {
				if f.Mode().IsRegular() {
					ext := strings.ToUpper(filepath.Ext(f.Name()))
					switch ext {
					// TODO: add other supported RAW files
					case ".NRW":
						fallthrough
					case ".CR2":
						data.Photos = append(data.Photos, Photo{Src: fullPath, Name: f.Name(), SupportedFormat: true, SendToBrowser: false, IsVideo: false})

					case ".MOV":
						ext := filepath.Ext(fullPath)
						thmPath := fullPath[0:len(fullPath)-len(ext)] + ".THM"

						data.Photos = append(data.Photos, Photo{Src: fullPath, Name: f.Name(),
							SupportedFormat: false,
							SendToBrowser:   false,
							IsVideo:         true,
							VideoPreview:    thmPath})

					case ".THM":
						// do nothing

					case "JPG":
						fallthrough
					case "JPEG":
						fallthrough
					case "JPE":
						fallthrough
					case "PNG":
						fallthrough
					case "GIF":
						fallthrough
					case "SVG":
						fallthrough
					case "SVGZ":
						data.Photos = append(data.Photos, Photo{Src: fullPath, Name: f.Name(), SupportedFormat: false, SendToBrowser: true, IsVideo: false})
					default:
						// TODO: try to get preview through libraw, oterwise fallback to default (send file as is)
						data.Photos = append(data.Photos, Photo{Src: fullPath, Name: f.Name(), SupportedFormat: true, SendToBrowser: false, IsVideo: false})

						//data.Photos = append(data.Photos, Photo{Src: fullPath, Name: f.Name(), SupportedFormat: false, SendToBrowser: true, IsVideo: false})
					}
				}
			}
		}

		return data, nil

	})

	data := item.Value().(PhotosPageData)

	return c.Render(http.StatusOK, "photos", data)
}

func process(decodedImageBuffer []byte, flip int) ([]byte, error) {
	//if flip == 0 {
	previewBuffer, err := bimg2.NewImage(decodedImageBuffer).Resize(300, 200)
	if err != nil {
		log.Printf("ERROR: %s", err)
	}
	return previewBuffer, err
	//} else {
	//
	//	image := bimg2.NewImage(decodedImageBuffer)
	//	options := bimg2.Options{
	//		Width:  300,
	//		Height: 200,
	//		Crop:   true,
	//
	//		Gravity: bimg2.GravitySmart,
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

	err := writeCacheHeaders(c, filePath)
	if err != nil {
		c.Logger().Errorf("Can't write cache headers: %s", err)
	}

	item, err := thumbnailsCache.Fetch(filePath, time.Hour*8, func() (interface{}, error) {
		rawProcessor := raw.NewRawProcessor()
		defer rawProcessor.Close()
		data, err := rawProcessor.ExtractPreview(filePath, process)
		if err != nil {
			return nil, err
		}

		return data, nil
	})
	if err != nil {
		return FetchHandler(c)
	}

	data := item.Value().([]byte)

	return c.Blob(http.StatusOK, "image/jpeg", data)
}

func processFull(decodedImageBuffer []byte, flip int) ([]byte, error) {
	// TODO: rotate photo if needed
	tmp := make([]byte, len(decodedImageBuffer))
	copy(tmp, decodedImageBuffer)
	return tmp, nil
	//
	//buf, err := bimg2.NewImage(decodedImageBuffer).SmartCrop(1920, 1200)
	//if err != nil {
	//	log.Printf("ERROR: %s", err)
	//}
	//return buf, err
}

func makeETagFromFile(fi os.FileInfo, filePath string) (string, error) {
	h := md5.New()
	_, err := io.WriteString(h, "ajo9e75thgalzdkfuhgei8a") // some salt
	if err != nil {
		panic("Can't make ETag")
	}

	int64ByteBuf := make([]byte, 2*binary.MaxVarintLen64)
	binary.PutVarint(int64ByteBuf, fi.ModTime().UTC().UnixNano())
	binary.PutVarint(int64ByteBuf[binary.MaxVarintLen64:2*binary.MaxVarintLen64], fi.Size())

	io.WriteString(h, filePath)

	// TODO: read some bytes from file for md5 sum?

	etag := h.Sum(int64ByteBuf)

	return hex.EncodeToString(etag), nil
}

func writeCacheHeaders(c echo.Context, filePath string) error {
	fi, err := os.Stat(filePath)
	if err != nil {
		c.Logger().Errorf("Error while making ETag header for file \"%s\": %s", filePath, err)
		return err
	}

	c.Response().Header().Set("Cache-Control", "max-age=3600")
	c.Response().Header().Set("Last-Modified", fi.ModTime().UTC().Format(http.TimeFormat))

	etagStr, err := makeETagFromFile(fi, filePath)
	if err != nil {
		c.Logger().Errorf("Error while making ETag header for file \"%s\": %s", filePath, err)
		return err
	} else {
		c.Response().Header().Set("ETag", etagStr)
	}

	return nil
}

func FetchHandler(c echo.Context) error {
	filePath := fullPath(c.QueryParam("s"))

	err := writeCacheHeaders(c, filePath)
	if err != nil {
		c.Logger().Errorf("Can't write cache headers: %s", err)
	}

	f, err := os.Open(filePath)
	if err != nil {
		c.Logger().Errorf("Can't open file \"%s\": %s", filePath, err)
		return echo.NotFoundHandler(c)
	}
	defer f.Close()

	contentType := mime.TypeByExtension(filepath.Ext(filePath))

	if contentType != "image/jpeg" {
		c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("%s; filename=%q", contentType, filepath.Base(filePath)))
	}

	return c.Stream(http.StatusOK, contentType, f)
}

func StreamVideoHandler(c echo.Context) error {
	filePath := fullPath(c.QueryParam("s"))

	// TODO: refactor extension hack (remove fake .mp4 which was placed for browser)
	ext := filepath.Ext(filePath)
	filePath = filePath[0 : len(filePath)-len(ext)]

	cmd := exec.Command("./bin/play-to-stdout.sh", filePath)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	// TODO: cmd.Wait() ?

	contentType := "mp4"
	//contentType := "video/x-flv"
	//contentType := "video/MP2T"

	c.Response().Header().Set("Accept-Ranges", "bytes")
	//videoFileName := filepath.Base(filePath) + ".mp4"
	//c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("%s; filename=%q", contentType, videoFileName))

	return c.Stream(200, contentType, stdout)
}

func FullPreviewPhotoHandler(c echo.Context) error {
	filePath := fullPath(c.QueryParam("s"))

	// TODO: move raw processor to worker pool
	rawProcessor := raw.NewRawProcessor()
	defer rawProcessor.Close()
	data, err := rawProcessor.ExtractPreview(filePath, processFull)
	if err != nil {
		//return err
		return FetchHandler(c)
	}

	err = writeCacheHeaders(c, filePath)
	if err != nil {
		c.Logger().Errorf("Can't write cache headers: %s", err)
	}

	return c.Blob(http.StatusOK, "image/jpeg", data)
}

func main() {
	// TODO: move magick numbers to default consts / env vars / config
	thumbnailsCache = ccache.New(ccache.Configure().MaxSize(10000).ItemsToPrune(500).GetsPerPromote(1))
	filesListCache = ccache.New(ccache.Configure().MaxSize(1000).ItemsToPrune(50).GetsPerPromote(1))

	// TODO: add another extensions
	mime.AddExtensionType(".CR2", "image/x-canon-cr2")
	mime.AddExtensionType(".NRW", "application/x-extension-NRW")

	if len(basePath()) < 1 {
		fmt.Fprintf(os.Stderr, "Please set path to gallery in GALLERY_PATH env var!")
		os.Exit(1)
	}

	tpl := &Template{
		templates: template.Must(template.ParseGlob("public/views/*.html")),
	}

	e := echo.New()
	e.Static("/css", "static/css")
	e.Static("/js", "static/js")
	e.Use(middleware.Logger())
	//e.Use(middleware.Recover())
	e.Renderer = tpl
	e.HideBanner = true

	e.GET("/", Index)
	e.GET("/d", Index)
	e.GET("/p", PreviewPhotoHandler)
	e.GET("/g", FullPreviewPhotoHandler)
	e.GET("/f", FetchHandler)
	e.GET("/v", StreamVideoHandler)
	e.Logger.Fatal(e.Start(":1323"))

}
