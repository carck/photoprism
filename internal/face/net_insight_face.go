//go:build INSIGHTFACE
// +build INSIGHTFACE

package face

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
)

// Net is a wrapper for the TensorFlow Facenet model.
type Net struct {
	modelPath string
	disabled  bool
}

type InsightFace struct {
	Box       []float64   `json:"bbox"`
	Kps       [][]float64 `json:"kps"`
	Score     float64     `json:"det_score"`
	Embedding []float64   `json:"embedding"`
}

// NewNet returns new TensorFlow instance with Facenet model.
func NewNet(modelPath string, cachePath string, disabled bool) *Net {
	return &Net{modelPath: modelPath, disabled: disabled}
}

// Detect runs the detection and facenet algorithms over the provided source image.
func (t *Net) Detect(fileName string, minSize int, cacheCrop bool, expected int) (faces Faces, err error) {
	var insightFaces []InsightFace
	resp, err := http.Get("http://localhost:8008?f=" + url.QueryEscape(fileName))

	if err != nil {
		return faces, err
	}

	err = json.NewDecoder(resp.Body).Decode(&insightFaces)
	if err != nil {
		return faces, err
	}

	for i := 0; i < len(insightFaces); i++ {
		w := insightFaces[i].Box[2] - insightFaces[i].Box[0]
		h := insightFaces[i].Box[3] - insightFaces[i].Box[1]

		imageWidth, _ := strconv.Atoi(resp.Header.Get("X-Width"))
		imageHeight, _ := strconv.Atoi(resp.Header.Get("X-Height"))

		faceCoord := NewArea(
			"face",
			int((insightFaces[i].Box[3]+insightFaces[i].Box[1])/2.0),
			int((insightFaces[i].Box[2]+insightFaces[i].Box[0])/2.0),
			int(Max64(w, h)),
		)
		faces = append(faces, Face{
			Rows:      imageHeight,
			Cols:      imageWidth,
			Score:     int(insightFaces[i].Score * 100),
			Area:      faceCoord,
			Eyes:      make([]Area, 0),
			Landmarks: make([]Area, 5),
		})
		for j := 0; j < 5; j++ {
			faces[i].Landmarks[j] = NewArea(
				"l",
				int(insightFaces[i].Kps[j][1]),
				int(insightFaces[i].Kps[j][0]),
				int(1),
			)
		}
		q, embedding := L2Norm64(insightFaces[i].Embedding, 1e-12)
		faces[i].Q = q
		faces[i].Embeddings = make(Embeddings, 1)
		faces[i].Embeddings[0] = NewEmbedding(embedding)
	}

	return faces, nil
}
