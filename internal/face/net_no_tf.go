//go:build LIBFACEDETECTION
// +build LIBFACEDETECTION

package face

import (
	"fmt"
	"image"
	"math"
	"path"
	"path/filepath"
	"runtime/debug"
	"sync"

	"github.com/carck/onnx-runtime-go"
	"github.com/photoprism/photoprism/internal/crop"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Net is a wrapper for the TensorFlow Facenet model.
type Net struct {
	model     *onnx.Model
	modelPath string
	disabled  bool
	modelName string
	modelTags []string
	mutex     sync.Mutex
	inFloats  []float32
}

// NewNet returns new TensorFlow instance with Facenet model.
func NewNet(modelPath string, cachePath string, disabled bool) *Net {
	return &Net{modelPath: modelPath, disabled: disabled, modelName: "facenet.onnx", modelTags: []string{"serve"}}
}

// Detect runs the detection and facenet algorithms over the provided source image.
func (t *Net) Detect(fileName string, minSize int, cacheCrop bool, expected int) (faces Faces, err error) {
	faces, err = Detect(fileName)

	if err != nil {
		return faces, err
	}

	if t.disabled {
		return faces, nil
	}

	err = t.loadModel()

	if err != nil {
		return faces, err
	}

	for i, f := range faces {
		if f.Area.Col == 0 && f.Area.Row == 0 {
			continue
		}
		q, embedding := t.getFaceEmbedding(fileName, f)

		if len(embedding) > 0 {
			faces[i].Q = q
			faces[i].Embeddings = make(Embeddings, 1)
			faces[i].Embeddings[0] = NewEmbedding(embedding)
		}
	}

	return faces, nil
}

// ModelLoaded tests if the TensorFlow model is loaded.
func (t *Net) ModelLoaded() bool {
	return t.model != nil
}

func (t *Net) loadModel() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.ModelLoaded() {
		return nil
	}

	modelPath := path.Join(t.modelPath, t.modelName)

	log.Infof("facenet: loading %s", txt.Quote(filepath.Base(modelPath)))

	shape := []int64{1, 3, 112, 112}
	inputNames := []string{"input.1"}
	outputNames := []string{"683"}

	model := onnx.NewModel(modelPath, shape, inputNames, outputNames, onnx.CPU)
	if model == nil {
		return fmt.Errorf("classify: load model failed, stack: %s", debug.Stack())
	}

	t.inFloats = make([]float32, 112*112*3)

	t.model = model

	return nil
}

func (t *Net) getFaceEmbedding(fileName string, f Face) (float64, []float32) {
	eyes := f.Eyes
	var img image.Image
	var err error

	if len(eyes) == 2 {
		x1 := float64(eyes[1].Col - eyes[0].Col)
		y1 := float64(eyes[1].Row - eyes[0].Row)

		angle := math.Atan2(y1, x1) * 180.0 / math.Pi

		img, err = crop.ImageFromThumb(fileName, f.CropArea(), CropSize, false, angle)
	} else {
		img, err = crop.ImageFromThumb(fileName, f.CropArea(), CropSize, false, 0)
	}

	if err != nil {
		log.Errorf("face: failed to crop image : %v", err)
		return 0, nil
	}

	err = t.imageToTensor(img, CropSize.Width, CropSize.Height)

	if err != nil {
		log.Errorf("face: failed to convert image to tensor: %v", err)
		return 0, nil
	}
	// TODO: prewhiten image as in facenet

	output := t.model.RunInference(t.inFloats)
	defer output.Delete()

	res := make([]float32, 512)
	output.CopyToBuffer(res, 512*4)
	norm := L2Norm(res, 1e-12)
	return norm, res
}

func (t *Net) imageToTensor(img image.Image, imageHeight, imageWidth int) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("face: %s (panic)\nstack: %s", r, debug.Stack())
		}
	}()

	if imageHeight <= 0 || imageWidth <= 0 {
		return fmt.Errorf("face: image width and height must be > 0")
	}

	ff := t.inFloats
	rs := imageHeight * imageWidth
	bs := 2 * rs
	for y := 0; y < imageHeight; y++ {
		for x := 0; x < imageWidth; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			base := y*imageWidth + x
			ff[base] = convertValue(r)
			ff[rs+base] = convertValue(g)
			ff[bs+base] = convertValue(b)
		}
	}
	return nil
}

func convertValue(value uint32) float32 {
	return (float32(value>>8) - float32(127.5)) / float32(127.5)
}
