package photoprism

import (
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/face"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/sanitize"
)

// Faces finds faces in JPEG media files and returns them.
func (ind *Index) Faces(jpeg *MediaFile, expected int) face.Faces {
	if jpeg == nil {
		return face.Faces{}
	}

	var thumbSize thumb.Name

	// Select best thumbnail depending on configured size.
	thumbSize = thumb.Fit1280

	thumbName, err := jpeg.Thumbnail(Config().ThumbPath(), thumbSize)

	if err != nil {
		log.Debugf("index: %s in %s (faces)", err, sanitize.Log(jpeg.BaseName()))
		return face.Faces{}
	}

	if thumbName == "" {
		log.Debugf("index: thumb %s not found in %s (faces)", thumbSize, sanitize.Log(jpeg.BaseName()))
		return face.Faces{}
	}

	start := time.Now()

	faces, err := ind.faceNet.Detect(thumbName, Config().FaceSize(), true, expected)

	if err != nil {
		log.Debugf("%s in %s", err, sanitize.Log(jpeg.BaseName()))
	}

	if l := len(faces); l > 0 {
		log.Infof("index: found %s in %s [%s]", english.Plural(l, "face", "faces"), sanitize.Log(jpeg.BaseName()), time.Since(start))
	}

	return faces
}
