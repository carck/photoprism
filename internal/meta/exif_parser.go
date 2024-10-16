package meta

import (
	"fmt"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/carck/libheif-go"
	"github.com/dsoprea/go-exif/v3"
	jpegstructure "github.com/dsoprea/go-jpeg-image-structure/v2"
	pngstructure "github.com/dsoprea/go-png-image-structure/v2"
	tiffstructure "github.com/dsoprea/go-tiff-image-structure/v2"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/sanitize"
)

func RawExif(fileName string, fileType fs.FileFormat, bruteForce bool) (rawExif []byte, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%s in %s (raw exif panic)\nstack: %s", e, sanitize.Log(filepath.Base(fileName)), debug.Stack())
		}
	}()

	// Extract raw Exif block.
	var parsed bool

	// Sanitized and shortened file name for logs.
	logName := sanitize.Log(filepath.Base(fileName))

	// Try Exif parser for specific media file format first.
	if fileType == fs.FormatJpeg {
		jpegMp := jpegstructure.NewJpegMediaParser()

		sl, err := jpegMp.ParseFile(fileName)

		if err != nil {
			log.Infof("metadata: %s while parsing jpeg file %s", err, logName)
		} else {
			_, rawExif, err = sl.Exif()

			if err != nil {
				if !bruteForce || strings.HasPrefix(err.Error(), "no exif header") {
					return rawExif, fmt.Errorf("found no exif header")
				} else if strings.HasPrefix(err.Error(), "no exif data") {
					log.Debugf("metadata: failed parsing %s, starting brute-force search (parse jpeg)", logName)
				} else {
					log.Infof("metadata: %s in %s, starting brute-force search (parse jpeg)", err, logName)
				}
			} else {
				parsed = true
			}
		}
	} else if fileType == fs.FormatPng {
		pngMp := pngstructure.NewPngMediaParser()

		cs, err := pngMp.ParseFile(fileName)

		if err != nil {
			return rawExif, fmt.Errorf("%s while parsing png file", err)
		} else {
			_, rawExif, err = cs.Exif()

			if err != nil {
				if err.Error() == "file does not have EXIF" || strings.HasPrefix(err.Error(), "no exif data") {
					return rawExif, fmt.Errorf("found no exif header")
				} else {
					log.Infof("metadata: %s in %s (parse png)", err, logName)
				}
			} else {
				parsed = true
			}
		}
	} else if fileType == fs.FormatHEIF {

		cs, err := heif.ReadExif(fileName)

		if err != nil {
			return rawExif, fmt.Errorf("%s while parsing heic file", err)
		} else {
			rawExif = cs

			parsed = true
		}
	} else if fileType == fs.FormatTiff {
		tiffMp := tiffstructure.NewTiffMediaParser()

		cs, err := tiffMp.ParseFile(fileName)

		if err != nil {
			return rawExif, fmt.Errorf("%s while parsing tiff file", err)
		} else {
			_, rawExif, err = cs.Exif()

			if err != nil {
				if err.Error() == "file does not have EXIF" || strings.HasPrefix(err.Error(), "no exif data") {
					return rawExif, fmt.Errorf("found no exif header")
				} else {
					log.Infof("metadata: %s in %s (parse tiff)", err, logName)
				}
			} else {
				parsed = true
			}
		}
	} else {
		log.Debugf("metadata: no native file format support for %s, performing brute-force exif search", logName)
		bruteForce = true
	}

	// Start brute-force search for Exif data?
	if !parsed && bruteForce {
		rawExif, err = exif.SearchFileAndExtractExif(fileName)

		if err != nil {
			return rawExif, fmt.Errorf("found no exif data")
		}
	}

	return rawExif, nil
}
