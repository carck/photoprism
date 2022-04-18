//go:build LIBFACEDETECTION
// +build LIBFACEDETECTION

package face

import (
	_ "embed"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/carck/libfacedetection-go"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

func init() {
	log.Infof("init face detector")
}

// Detect runs the detection algorithm over the provided source image.
func Detect(fileName string) (faces Faces, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("faces: %s (panic)\nstack: %s", r, debug.Stack())
		}
	}()

	if !fs.FileExists(fileName) {
		return faces, fmt.Errorf("faces: file '%s' not found", txt.Quote(filepath.Base(fileName)))
	}

	faces, err = LibFaceDetectionDetect(fileName)

	if err != nil {
		return faces, fmt.Errorf("faces: %s", err)
	}

	return faces, nil
}

// Detect runs the detection algorithm over the provided source image.
func LibFaceDetectionDetect(fileName string) (results Faces, err error) {
	r, err := os.Open(fileName)
	if err != nil {
		log.Errorf("faces: %s", err)
		return nil, err
	}
	defer r.Close()

	m, _, err := image.Decode(r)
	if err != nil {
		log.Errorf("faces: %s", err)
		return nil, err
	}

	rgb, w, h := libfacedetection.NewRGBImageFrom(m)

	faces := libfacedetection.DetectFaceRGB(rgb, w, h, w*3)

	for i := 0; i < len(faces); i++ {
		var eyesCoords []Area
		var landmarkCoords []Area

		if faces[i].Confidence < 70 {
			continue
		}

		q := faces[i].W
		if faces[i].H > q {
			q = faces[i].H
		}
		faceCoord := NewArea(
			"face",
			int(faces[i].Y+faces[i].H/2.0),
			int(faces[i].X+faces[i].W/2.0),
			q,
		)

		for j := 0; j < 5; j++ {
			if j < 2 {
				eyesCoords = append(eyesCoords, NewArea(
					"eye",
					faces[i].Landmarks[j*2+1],
					faces[i].Landmarks[j*2],
					int(1),
				))
			} else {
				landmarkCoords = append(landmarkCoords, NewArea(
					"mouth",
					faces[i].Landmarks[j*2+1],
					faces[i].Landmarks[j*2],
					int(1),
				))
			}
		}

		results = append(results, Face{
			Rows:      h,
			Cols:      w,
			Score:     int(faces[i].Confidence),
			Area:      faceCoord,
			Eyes:      eyesCoords,
			Landmarks: landmarkCoords,
		})
	}

	return results, nil
}
