package face

import (
	"bytes"
	"encoding/json"

	"github.com/montanaflynn/stats"
	"github.com/photoprism/photoprism/pkg/clusters"
	"github.com/valyala/gozstd"
)

// Embeddings represents a face embedding cluster.
type Embeddings []Embedding

// NewEmbeddings creates a new embeddings from inference results.
func NewEmbeddings(inference [][]float32) Embeddings {
	result := make(Embeddings, len(inference))

	var v []float32
	var i int

	for i, v = range inference {
		e := NewEmbedding(v)

		if e.NotBlacklisted() {
			result[i] = e
		}
	}

	return result
}

// Empty tests if embeddings are empty.
func (embeddings Embeddings) Empty() bool {
	if len(embeddings) < 1 {
		return true
	}

	return len(embeddings[0]) < 1
}

// Count returns the number of embeddings.
func (embeddings Embeddings) Count() int {
	if embeddings.Empty() {
		return 0
	}

	return len(embeddings)
}

// One tests if there is exactly one embedding.
func (embeddings Embeddings) One() bool {
	return embeddings.Count() == 1
}

// First returns the first face embedding.
func (embeddings Embeddings) First() Embedding {
	if embeddings.Empty() {
		return NullEmbedding
	}

	return embeddings[0]
}

// Float32 returns embeddings as a float32 slice.
func (embeddings Embeddings) Float32() [][]float32 {
	result := make([][]float32, len(embeddings))

	for i, e := range embeddings {
		result[i] = e
	}

	return result
}

// Contains tests if another embeddings is contained within a radius.
func (embeddings Embeddings) Contains(other Embedding, radius float64) bool {
	for _, e := range embeddings {
		if d := e.Distance(other); d < radius {
			return true
		}
	}

	return false
}

// Distance returns the minimum distance to an embedding.
func (embeddings Embeddings) Distance(other Embedding) (dist float64) {
	dist = -1

	for _, e := range embeddings {
		if d := e.Distance(other); d < dist || dist < 0 {
			dist = d
		}
	}

	return dist
}

// JSON returns the embeddings as JSON bytes.
func (embeddings Embeddings) JSON() []byte {
	var noResult = []byte("")

	if embeddings.Empty() {
		return noResult
	}

	if result, err := json.Marshal(embeddings); err != nil {
		return noResult
	} else {
		return gozstd.Compress(nil, result)
	}
}

// EmbeddingsMidpoint returns the embeddings vector midpoint.
func EmbeddingsMidpoint(embeddings Embeddings) (result Embedding, radius float64, count int) {
	// Return if there are no embeddings.
	if embeddings.Empty() {
		return Embedding{}, 0, 0
	}

	// Count embeddings.
	count = len(embeddings)

	// Only one embedding?
	if count == 1 {
		// Return embedding if there is only one.
		return embeddings[0], 0.35, 1
	}

	dim := len(embeddings[0])

	// No embedding values?
	if dim == 0 {
		return Embedding{}, 0.0, count
	}

	result = make(Embedding, dim)

	// The mean of a set of vectors is calculated component-wise.
	for i := 0; i < dim; i++ {
		values := make(stats.Float64Data, count)

		for j := 0; j < count; j++ {
			values[j] = float64(embeddings[j][i])
		}

		if m, err := stats.Mean(values); err != nil {
			log.Warnf("embeddings: %s", err)
		} else {
			result[i] = float32(m)
		}
	}

	// Radius is the max embedding distance + 0.01 from result.
	for _, emb := range embeddings {
		if d := clusters.EuclideanDistance32(result, emb); d > radius {
			radius = d + 0.01
		}
	}

	return result, radius, count
}

// UnmarshalEmbeddings parses face embedding JSON.
func UnmarshalEmbeddings(s []byte) (result Embeddings) {
	if decompressed, err := gozstd.Decompress(nil, s); err != nil {
		log.Errorf("faces: decompress %s", err)
	} else {
		if !bytes.HasPrefix(decompressed, []byte("[[")) {
			return nil
		}

		if err := json.Unmarshal(decompressed, &result); err != nil {
			log.Errorf("faces: %s", err)
		}
	}

	return result
}
