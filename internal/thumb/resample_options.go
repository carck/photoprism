package thumb

import (
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/imaging"
)

type ResampleOption int

const (
	ResampleFillCenter ResampleOption = iota
	ResampleFillTopLeft
	ResampleFillBottomRight
	ResampleFit
	ResampleResize
	ResampleNearestNeighbor
	ResampleDefault
	ResamplePng
)

var ResampleMethods = map[ResampleOption]string{
	ResampleFillCenter:      "center",
	ResampleFillTopLeft:     "left",
	ResampleFillBottomRight: "right",
	ResampleFit:             "fit",
	ResampleResize:          "resize",
}

// ResampleOptions extracts filter, format, and method from resample options.
func ResampleOptions(opts ...ResampleOption) (method ResampleOption, filter imaging.ResampleFilter, format fs.FileFormat) {
	method = ResampleFit
	filter = imaging.Lanczos
	format = fs.FormatJpeg

	for _, option := range opts {
		switch option {
		case ResamplePng:
			format = fs.FormatPng
		case ResampleNearestNeighbor:
			filter = imaging.NearestNeighbor
		case ResampleDefault:
			filter = Filter.Imaging()
		case ResampleFillTopLeft:
			method = ResampleFillTopLeft
		case ResampleFillCenter:
			method = ResampleFillCenter
		case ResampleFillBottomRight:
			method = ResampleFillBottomRight
		case ResampleFit:
			method = ResampleFit
		case ResampleResize:
			method = ResampleResize
		}
	}

	return method, filter, format
}
