//go:build NOTENSORFLOW
// +build NOTENSORFLOW

package classify

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	"io/ioutil"
	"math"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"
	"sync"

	"github.com/disintegration/imaging"
	"github.com/mattn/go-tflite"
	"github.com/mattn/go-tflite/delegates/xnnpack"
	"github.com/photoprism/photoprism/pkg/txt"
)

// TensorFlow is a wrapper for tensorflow low-level API.
type TensorFlow struct {
	mu          sync.Mutex
	interpreter *tflite.Interpreter
	modelsPath  string
	disabled    bool
	modelName   string
	modelFile   string
	labels      []string
	ignores     map[string]bool
	inBytes     []byte
	inFloats    []float32
	outFloats   []float32
}

// New returns new TensorFlow instance with Nasnet model.
func New(modelsPath string, disabled bool) *TensorFlow {
	return &TensorFlow{modelsPath: modelsPath, disabled: disabled, modelName: "mobile_ica", modelFile: "mobile_ica.tflite"}
}

// Init initialises tensorflow models if not disabled
func (t *TensorFlow) Init() (err error) {
	if t.disabled {
		return nil
	}

	return t.loadModel()
}

// File returns matching labels for a jpeg media file.
func (t *TensorFlow) File(filename string) (result Labels, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.disabled {
		return result, nil
	}

	imageBuffer, err := ioutil.ReadFile(filename)

	if err != nil {
		return nil, err
	}

	return t.Labels(imageBuffer)
}

// Labels returns matching labels for a jpeg media string.
func (t *TensorFlow) Labels(img []byte) (result Labels, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("classify: %s (inference panic)\nstack: %s", r, debug.Stack())
		}
	}()

	if t.disabled {
		return result, nil
	}

	if err := t.loadModel(); err != nil {
		return nil, err
	}

	// Create tensor from image.
	err = t.createTensor(img, "jpeg")

	if err != nil {
		return nil, err
	}

	// Run inference.
	status := t.interpreter.Invoke()
	if status != tflite.OK {
		return result, fmt.Errorf("classify: %s (run inference)", err.Error())
	}

	output := t.interpreter.GetOutputTensor(0)

	var scores []float32
	if output.Type() == tflite.Float32 {
		scores = output.Float32s()
	} else if output.Type() == tflite.UInt8 {
		output_size := output.Dim(output.NumDims() - 1)
		scores = t.outFloats
		outBytes := output.UInt8s()
		for i := 0; i < output_size; i++ {
			scores[i] = float32(float64(outBytes[i]) / 255.0)
		}
	}
	// Return best labels
	result = t.bestLabels(scores)
	if len(result) > 0 {
		log.Tracef("classify: image classified as %+v", result)
	}

	return result, nil
}

func (t *TensorFlow) loadLabels(path string, file string) ([]string, error) {
	modelLabels := path + "/" + file

	log.Infof("classify: loading labels from labels.txt")

	// Load labels
	f, err := os.Open(modelLabels)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)

	var labels []string
	// Labels are separated by newlines
	for scanner.Scan() {
		labels = append(labels, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return labels, nil
}

// ModelLoaded tests if the TensorFlow model is loaded.
func (t *TensorFlow) ModelLoaded() bool {
	return t.interpreter != nil
}

func (t *TensorFlow) loadModel() error {
	if t.ModelLoaded() {
		return nil
	}

	modelPath := path.Join(t.modelsPath, t.modelName)

	log.Infof("classify: loading %s", txt.Quote(filepath.Base(modelPath)))

	model := tflite.NewModelFromFile(path.Join(modelPath, t.modelFile))
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
	inType := input.Type()

	if inType == tflite.UInt8 {
		t.inBytes = make([]byte, h*w*c)
	} else if inType == tflite.Float32 {
		t.inFloats = make([]float32, h*w*c)
	} else {
		return fmt.Errorf("is not wanted type")
	}

	output := interpreter.GetOutputTensor(0)
	if output.Type() == tflite.UInt8 {
		output_size := output.Dim(output.NumDims() - 1)
		t.outFloats = make([]float32, output_size)
	}

	t.interpreter = interpreter

	if labels, err := t.loadLabels(modelPath, "labels.txt"); err != nil {
		return err
	} else {
		t.labels = labels
	}

	if labels, err := t.loadLabels(modelPath, "ignores.txt"); err != nil {
		return err
	} else {
		for _, element := range labels {
			t.ignores[element] = true
		}
	}
	return nil
}

// bestLabels returns the best 5 labels (if enough high probability labels) from the prediction of the model
func (t *TensorFlow) bestLabels(probabilities []float32) Labels {
	var result Labels

	for i, p := range probabilities {
		if i >= len(t.labels) {
			// break if probabilities and labels does not match
			break
		}

		// discard labels with low probabilities
		if p < 0.82 {
			continue
		}

		labelText := strings.ToLower(t.labels[i])

		if _, ok := t.ignores[labelText]; ok {
			continue
		}

		uncertainty := 100 - int(math.Round(float64(p*100)))

		result = append(result, Label{Name: labelText, Source: SrcImage, Uncertainty: uncertainty, Priority: 1})
	}

	// Sort by probability
	sort.Sort(result)

	// Return the best labels only.
	if l := len(result); l < 5 {
		return result[:l]
	} else {
		return result[:5]
	}
}

// createTensor converts bytes jpeg image in a tensor object required as tensorflow model input
func (t *TensorFlow) createTensor(image []byte, imageFormat string) error {
	img, err := imaging.Decode(bytes.NewReader(image), imaging.AutoOrientation(true))

	if err != nil {
		return err
	}

	width, height := 224, 224

	img = imaging.Fill(img, width, height, imaging.Center, imaging.Lanczos)

	return t.imageToTensor(img, width, height)
}

func (t *TensorFlow) imageToTensor(img image.Image, imageHeight, imageWidth int) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("classify: %s (panic)\nstack: %s", r, debug.Stack())
		}
	}()

	if imageHeight <= 0 || imageWidth <= 0 {
		return fmt.Errorf("classify: image width and height must be > 0")
	}

	input := t.interpreter.GetInputTensor(0)
	wanted_height := input.Dim(1)
	wanted_width := input.Dim(2)
	wanted_type := input.Type()

	if wanted_type == tflite.UInt8 {
		bb := t.inBytes
		for y := 0; y < wanted_height; y++ {
			for x := 0; x < wanted_width; x++ {
				col := img.At(x, y)
				r, g, b, _ := col.RGBA()
				bb[(y*wanted_width+x)*3+0] = byte(float64(r) / 255.0)
				bb[(y*wanted_width+x)*3+1] = byte(float64(g) / 255.0)
				bb[(y*wanted_width+x)*3+2] = byte(float64(b) / 255.0)
			}
		}
		input.CopyFromBuffer(bb)
	} else if wanted_type == tflite.Float32 {
		ff := t.inFloats
		for y := 0; y < wanted_height; y++ {
			for x := 0; x < wanted_width; x++ {
				r, g, b, _ := img.At(x, y).RGBA()
				ff[(y*wanted_width+x)*3+0] = float32(r) / 65535.0
				ff[(y*wanted_width+x)*3+1] = float32(g) / 65535.0
				ff[(y*wanted_width+x)*3+2] = float32(b) / 65535.0
			}
		}
		copy(input.Float32s(), ff)
	} else {
		return fmt.Errorf("is not wanted type")
	}

	return nil
}

func convertValue(value uint32) float32 {
	return (float32(value>>8) - float32(127.5)) / float32(127.5)
}
