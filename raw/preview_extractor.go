package raw

/*
#cgo LDFLAGS: -lraw_r -lgomp
#include <libraw/libraw.h>
*/
import "C"
import (
	"fmt"
	"unsafe"
)

type PreviewOptions struct {
	Width  int
	Height int
}

func panicOnError(filePath string, result int) {
	if result != 0 {
		strError := C.GoString(C.libraw_strerror(C.int(result)))
		panic(fmt.Sprintf("LibRaw error for file \"%s\": %s (%d)", filePath, strError, result))
	}
}

type Raw struct {
	libRaw *C.libraw_data_t
}

func NewRawProcessor() *Raw {
	r := Raw{}
	r.init()

	return &r
}

func (r *Raw) init() {
	r.libRaw = (*C.libraw_data_t)(unsafe.Pointer(C.libraw_init(0)))
}

func (r *Raw) Close() {
	C.libraw_close(r.libRaw)
	//C.free(unsafe.Pointer(r.libRaw))
}

func (r *Raw) ExtractPreview(filePath string, callback func(decodedImageBuffer []byte, flip int) ([]byte, error)) []byte {
	cPath := C.CString(filePath)
	defer C.free(unsafe.Pointer(cPath))
	resultCode := int(C.libraw_open_file(r.libRaw, cPath))

	panicOnError(filePath, resultCode)

	resultCode = int(C.libraw_unpack_thumb(r.libRaw))
	panicOnError(filePath, resultCode)

	defer C.libraw_recycle(r.libRaw)

	// TODO: if (iProcessor.imgdata.thumbnail.tformat==LIBRAW_THUMBNAIL_JPEG) else if (iProcessor.imgdata.thumbnail.tformat==LIBRAW_THUMBNAIL_BITMAP)
	var buffer []byte
	const buf_size = 1 << 30
	if r.libRaw.thumbnail.tlength > buf_size {
		panic(fmt.Sprintf("Too big thumbnail image \"%s\" %d bytes. It larger than limit of %d bytes", filePath, int(r.libRaw.thumbnail.tlength), buf_size))
	}
	buffer = (*[buf_size]byte)(unsafe.Pointer(r.libRaw.thumbnail.thumb))[0:r.libRaw.thumbnail.tlength]

	newImage, err := callback(buffer, int(r.libRaw.sizes.flip))
	if err != nil {
		panic(err) // TODO: return errors
	}

	return newImage
}
