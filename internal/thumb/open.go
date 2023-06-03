package thumb

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/carck/libheif/go/heif"
	"github.com/disintegration/imaging"
	"github.com/mandykoh/prism/meta/icc"

	"github.com/photoprism/photoprism/pkg/colors"
	"github.com/photoprism/photoprism/pkg/fs"
)

// StandardRGB configures whether colors in the Apple Display P3 color space should be converted to standard RGB.
var StandardRGB = true

// Open loads an image from disk, rotates it, and converts the color profile if necessary.
func Open(fileName string, orientation int) (result image.Image, err error) {
	if fileName == "" {
		return result, fmt.Errorf("filename missing")
	}

	// Open JPEG?
	if StandardRGB && fs.GetFileFormat(fileName) == fs.FormatJpeg {
		return OpenJpeg(fileName, orientation)
	}

	if fs.GetFileFormat(fileName) == fs.FormatHEIF {
		if result, err = OpenHeif(fileName); err == nil {
			return result, nil
		}
	}

	// Open file with imaging function.
	img, err := imaging.Open(fileName)

	if err != nil {
		return result, err
	}

	// Rotate?
	if orientation > 1 {
		img = Rotate(img, orientation)
	}

	return img, nil
}

func OpenHeif(fileName string) (image.Image, error) {
	c, err := heif.NewContext()
	if err != nil {
		return nil, err
	}
	if err := c.ReadFromFile(fileName); err != nil {
		return nil, err
	}
	handle, err := c.GetPrimaryImageHandle()
	if err != nil {
		return nil, err
	}

	img, err := handle.DecodeImage(heif.ColorspaceUndefined, heif.ChromaUndefined, nil)
	if err != nil {
		return nil, err
	}

	image, err := img.GetImage()
	if err != nil {
		return nil, err
	}

	if data, _ := handle.GetICCProfle(); data != nil {
		md, _ := icc.NewProfileReader(bytes.NewReader(data)).ReadProfile()
		profile, _ := md.Description()
		switch {
		case colors.ProfileDisplayP3.Equal(profile):
			image = colors.ToSRGB(image, colors.ProfileDisplayP3)
		}
	}
	return image, nil
}
