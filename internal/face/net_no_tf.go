// +build NOTENSORFLOW

package face

import (
	"bytes"
	"fmt"
	"image"
	"io/ioutil"
	"math"
	"path"
	"path/filepath"
	"runtime/debug"
	"sync"

	"github.com/carck/gg"
	"github.com/disintegration/imaging"
	"github.com/mattn/go-tflite"
	"github.com/mattn/go-tflite/delegates/xnnpack"
	"github.com/photoprism/photoprism/pkg/txt"
	"golang.org/x/image/draw"
)

// Net is a wrapper for the TensorFlow Facenet model.
type Net struct {
	interpreter *tflite.Interpreter
	modelPath   string
	disabled    bool
	modelName   string
	modelTags   []string
	mutex       sync.Mutex
	inFloats    []float32
}

// NewNet returns new TensorFlow instance with Facenet model.
func NewNet(modelPath string, cachePath string, disabled bool) *Net {
	return &Net{modelPath: modelPath, disabled: disabled, modelName: "mobile_facenet.tflite", modelTags: []string{"serve"}}
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

		embedding := t.getFaceEmbedding(fileName, f.Area, f.Eyes)

		if len(embedding) > 0 {
			faces[i].Embeddings = make([][]float32, 1)
			faces[i].Embeddings[0] = embedding
		}
	}

	return faces, nil
}

// ModelLoaded tests if the TensorFlow model is loaded.
func (t *Net) ModelLoaded() bool {
	return t.interpreter != nil
}

func (t *Net) loadModel() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.ModelLoaded() {
		return nil
	}

	modelPath := path.Join(t.modelPath, t.modelName)

	log.Infof("facenet: loading %s", txt.Quote(filepath.Base(modelPath)))

	model := tflite.NewModelFromFile(modelPath)
	if model == nil {
		return fmt.Errorf("classify: load model failed, stack: %s", debug.Stack())
	}

	options := tflite.NewInterpreterOptions()
	options.AddDelegate(xnnpack.New(xnnpack.DelegateOptions{NumThreads: 2}))
	options.SetNumThread(4)
	options.SetErrorReporter(func(msg string, user_data interface{}) {
		fmt.Println(msg)
	}, nil)

	interpreter := tflite.NewInterpreter(model, options)
	if interpreter == nil {
		defer options.Delete()
		defer model.Delete()
		return fmt.Errorf("classify: create interceptor failed, stack: %s", debug.Stack())
	}

	status := interpreter.AllocateTensors()
	if status != tflite.OK {
		defer interpreter.Delete()
		defer options.Delete()
		defer model.Delete()
		return fmt.Errorf("classify: create tensor failed, stack: %s", debug.Stack())
	}

	input := interpreter.GetInputTensor(0)
	h := input.Dim(1)
	w := input.Dim(2)
	c := input.Dim(3)
	t.inFloats = make([]float32, h*w*c)

	log.Infof("facenet: input %d %d %d", h, w, c)

	t.interpreter = interpreter

	return nil
}

func (t *Net) getFaceEmbedding(fileName string, f Area, eyes Areas) []float32 {
	y, x := f.TopLeft()

	imageBuffer, err := ioutil.ReadFile(fileName)
	img, err := imaging.Decode(bytes.NewReader(imageBuffer), imaging.AutoOrientation(true))
	if err != nil {
		log.Errorf("face: failed to decode image: %v", err)
	}

	if len(eyes) == 2 {
		x1 := float64(eyes[1].Col - eyes[0].Col)
		y1 := float64(eyes[1].Row - eyes[0].Row)

		angle := math.Atan2(y1, x1) * 180.0 / math.Pi

		dc := gg.NewContext(112, 112)
		dc.SetRGB255(255, 255, 255)
		dc.Clear()

		dc.RotateAbout(gg.Radians(-angle), 56, 56)
		dc.Scale(112/float64(f.Scale), 112/float64(f.Scale))

		dc.DrawImageAnchoredWithTransformer(img, 0, 0, float64(x)/float64(img.Bounds().Dx()), float64(y)/float64(img.Bounds().Dy()), draw.CatmullRom)
		img = dc.Image()

		//dc.SavePNG(path.Join("/home/l2/face", fmt.Sprintf("%s.png", f.String())))

	} else {
		img = imaging.Crop(img, image.Rect(x, y, x+f.Scale, y+f.Scale))
		img = imaging.Fill(img, 112, 112, imaging.Center, imaging.Lanczos)
	}

	err = t.imageToTensor(img, 112, 112)

	if err != nil {
		log.Errorf("face: failed to convert image to tensor: %v", err)
	}
	// TODO: prewhiten image as in facenet

	// Run inference.
	status := t.interpreter.Invoke()
	if status != tflite.OK {
		log.Errorf("face: failed to invoke: %v", err)
		return nil
	}

	output := t.interpreter.GetOutputTensor(0)
	outSize := output.Dim(output.NumDims() - 1)
	log.Infof("facenet: output %d", outSize)

	if outSize < 1 {
		log.Errorf("face: inference failed, no output")
	} else {
		result := make([]float32, outSize)
		copy(result, output.Float32s())
		return result
	}
	return nil
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

	input := t.interpreter.GetInputTensor(0)
	ff := t.inFloats
	for y := 0; y < imageHeight; y++ {
		for x := 0; x < imageWidth; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			ff[(y*imageWidth+x)*3+0] = convertValue(r)
			ff[(y*imageWidth+x)*3+1] = convertValue(g)
			ff[(y*imageWidth+x)*3+2] = convertValue(b)
		}
	}
	copy(input.Float32s(), ff)
	return nil
}

func convertValue(value uint32) float32 {
	return (float32(value>>8) - float32(127.5)) / float32(127.5)
}
