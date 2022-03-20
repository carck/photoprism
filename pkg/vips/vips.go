package vips

/*
#cgo pkg-config: vips
#include "vips.h"
*/
import "C"

import (
	"fmt"
	"unsafe"
)

const (
	InterestingNone      int = C.VIPS_INTERESTING_NONE
	InterestingCentre    int = C.VIPS_INTERESTING_CENTRE
	InterestingEntropy   int = C.VIPS_INTERESTING_ENTROPY
	InterestingAttention int = C.VIPS_INTERESTING_ATTENTION
	InterestingLow       int = C.VIPS_INTERESTING_LOW
	InterestingHigh      int = C.VIPS_INTERESTING_HIGH
	InterestingAll       int = C.VIPS_INTERESTING_ALL
	InterestingLast      int = C.VIPS_INTERESTING_LAST
)

func Start() {
	err := C.vips_init(C.CString("photoprism"))
	if err != 0 {
		panic("unable to start vips!")
	}
	C.vips_concurrency_set(1)
}

func Shutdown() {
	C.vips_shutdown()
}

func Thumbnail(fileName, thumbnailFileName string, width, height, crop, q int) error {
	cFileName := C.CString(fileName)
	defer C.free(unsafe.Pointer(cFileName))

	cOutName := C.CString(thumbnailFileName)
	defer C.free(unsafe.Pointer(cOutName))

	err := C.thumbnail(cFileName,
		cOutName,
		C.int(width),
		C.int(height),
		C.int(crop),
		C.int(q))
	if err != 0 {
		return handleVipsError()
	}
	return nil
}

func handleVipsError() error {
	s := C.GoString(C.vips_error_buffer())
	C.vips_error_clear()

	return fmt.Errorf("%v\n", s)
}
