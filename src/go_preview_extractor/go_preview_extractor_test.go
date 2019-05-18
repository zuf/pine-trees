package go_preview_extractor

import (
	"fmt"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
	"gopkg.in/h2non/bimg.v1"
	"io/ioutil"
	"os"

	"testing"

	"github.com/zuf/pine-trees/src/src/raw"
)

const testFile = "../../RAW-test/20140921_124657_IMG_9102.CR2"

func BenchmarkOnlyDecode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		f, err := os.Open(testFile)
		defer f.Close() // TODO: should verify errors on file close

		if err != nil {
			panic(err)
		}

		// Optionally register camera makenote data parsing - currently Nikon and
		// Canon are supported.
		exif.RegisterParsers(mknote.All...)

		_, err = exif.Decode(f)
		if err != nil {
			panic(err)
		}

		//bytesBuf, err := x.JpegThumbnail()
	}
}

func BenchmarkOnlyDecodeAndExtractJPEG(b *testing.B) {
	for i := 0; i < b.N; i++ {
		f, err := os.Open(testFile)
		defer f.Close() // TODO: should verify errors on file close

		if err != nil {
			panic(err)
		}

		// Optionally register camera makenote data parsing - currently Nikon and
		// Canon are supported.
		exif.RegisterParsers(mknote.All...)

		x, err := exif.Decode(f)
		if err != nil {
			panic(err)
		}

		buf, err := x.JpegThumbnail()
		if err != nil {
			panic(err)
		}

		tag, err := x.Get(mknote.Preview)
		if err != nil {
			panic(err)
		}

		preview := tag.Val

		fmt.Printf("Size of thumb: %d bytes\n", len(buf))
		err = ioutil.WriteFile("/tmp/test_thumb.jpg", buf, 0644)

		fmt.Printf("Size of preview: %d bytes\n", len(preview))
		err = ioutil.WriteFile("/tmp/test_preview.jpg", preview, 0644)

	}
}

func BenchmarkExtractFromExifAndProcess(b *testing.B) {
	for i := 0; i < b.N; i++ {
		buf, err := JPEGPreviewFromExif(testFile)
		if err != nil {
			panic(err)
		}
		process(buf, 0)
	}
}

//func BenchmarkLibRaw(b *testing.B) {
//	r := raw.NewRawProcessor()
//	defer r.Close()
//
//	for i := 0; i < b.N; i++ {
//		r.ExtractPreview(testFile, process)
//	}
//}

func process(decodedImageBuffer []byte, flip int) ([]byte, error) {
	//if flip == 0 {
	previewBuffer, err := bimg.NewImage(decodedImageBuffer).Resize(300, 200)
	if err != nil {
		panic(err)
	}
	return previewBuffer, err
}

//func BenchmarkDecodeThumbParallel(b *testing.B) {
//	b.RunParallel(func(pb *testing.PB) {
//		r := raw.NewRawProcessor()
//		defer r.Close()
//		for pb.Next() {
//			DecodeThumb(r, "./samples/TEST.CR2")
//		}fmt
//	})
//}
