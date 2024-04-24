package thumb

import (
	"github.com/carck/vips-thumbnail-go"
)

// Resample downscales an image and returns it.
func ResampleVips(soure, fileName string, width, height int, opts ...ResampleOption) (err error) {
	method, _, _ := ResampleOptions(opts...)

	q := int(JpegQuality)

	if width <= 150 && height <= 150 {
		q = int(JpegQualitySmall)
	}

	if method == ResampleFit {
		err = vips.ThumbnailDefault(soure, fileName, width, height, -1, q)
	} else if method == ResampleFillCenter {
		err = vips.ThumbnailDefault(soure, fileName, width, height, vips.InterestingCentre, q)
	} else if method == ResampleFillTopLeft {
		err = vips.ThumbnailDefault(soure, fileName, width, height, vips.InterestingLow, q)
	} else if method == ResampleFillBottomRight {
		err = vips.ThumbnailDefault(soure, fileName, width, height, vips.InterestingHigh, q)
	} else if method == ResampleResize {
		err = vips.ThumbnailDefault(soure, fileName, width, height, vips.InterestingCentre, q)
	}

	return err
}
