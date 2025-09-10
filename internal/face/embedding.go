package face

import (
	"encoding/json"
	"strings"

	"github.com/photoprism/photoprism/pkg/clusters"
)

// Embedding represents a face embedding.
type Embedding []float32

var NullEmbedding = make(Embedding, 512)

// NewEmbedding creates a new embedding from an inference result.
func NewEmbedding(inference []float32) Embedding {
	return inference
}

// Blacklisted tests if the face embedding is blacklisted.
func (m Embedding) Blacklisted() bool {
	return Blacklist.Contains(m, BlacklistRadius)
}

// Child tests if the face embedding belongs to a child.
func (m Embedding) Child() bool {
	return Children.Contains(m, ChildrenRadius)
}

// Unsuitable tests if the face embedding is unsuitable for clustering and matching.
func (m Embedding) Unsuitable() bool {
	return m.Child() || m.Blacklisted()
}

// Distance calculates the distance to another face embedding.
func (m Embedding) Distance(other Embedding) float64 {
	return clusters.EuclideanDistance32(m, other)
}

// Magnitude returns the face embedding vector length (magnitude).
func (m Embedding) Magnitude() float64 {
	return m.Distance(NullEmbedding)
}

// NotBlacklisted tests if the face embedding is not blacklisted.
func (m Embedding) NotBlacklisted() bool {
	return !m.Blacklisted()
}
