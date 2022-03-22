package thumb

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"

	"github.com/carck/libheif/go/heif"
	"github.com/disintegration/imaging"
	"github.com/mandykoh/prism/meta/autometa"
	"github.com/mandykoh/prism/meta/icc"

	"github.com/photoprism/photoprism/pkg/colors"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/sanitize"
)

// Open loads an image from disk, rotates it, and converts the color profile if necessary.
func Open(fileName string, orientation int) (result image.Image, err error) {
	if fileName == "" {
		return result, fmt.Errorf("filename missing")
	}

	// Open JPEG?
	if fs.GetFileFormat(fileName) == fs.FormatJpeg {
		return OpenJpeg(fileName, orientation)
	}

	if fs.GetFileFormat(fileName) == fs.FormatHEIF {
		return OpenHeif(fileName)
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

// OpenJpeg loads a JPEG image from disk, rotates it, and converts the color profile if necessary.
func OpenJpeg(fileName string, orientation int) (result image.Image, err error) {
	if fileName == "" {
		return result, fmt.Errorf("filename missing")
	}

	logName := sanitize.Log(filepath.Base(fileName))

	// Open file.
	fileReader, err := os.Open(fileName)

	if err != nil {
		return result, err
	}

	defer fileReader.Close()

	// Read color metadata.
	md, imgStream, err := autometa.Load(fileReader)

	var img image.Image

	if err != nil {
		log.Warnf("resample: %s in %s (read color metadata)", err, logName)
		img, err = imaging.Decode(fileReader)
	} else {
		img, err = imaging.Decode(imgStream)
	}

	if err != nil {
		return result, err
	}

	// Read ICC profile and convert colors if possible.
	if md != nil {
		if iccProfile, err := md.ICCProfile(); err != nil || iccProfile == nil {
			// Do nothing.
			log.Tracef("resample: %s has no color profile", logName)
		} else if profile, err := iccProfile.Description(); err == nil && profile != "" {
			log.Tracef("resample: %s has color profile %s", logName, sanitize.Log(profile))
			switch {
			case colors.ProfileDisplayP3.Equal(profile):
				img = colors.ToSRGB(img, colors.ProfileDisplayP3)
			}
		}
	}

	// Rotate?
	if orientation > 1 {
		img = Rotate(img, orientation)
	}

	return img, nil
}
