package photoprism

import "github.com/photoprism/photoprism/internal/face"
import "github.com/photoprism/photoprism/internal/config"

type FacesOptions struct {
	Force     bool
	Threshold int
}

// SampleThreshold returns the face embeddings sample threshold for clustering.
func (o FacesOptions) SampleThreshold() int {
	if o.Threshold > 0 {
		return o.Threshold
	}

	// Return default.
	return face.SampleThreshold
}

// FacesOptionsDefault returns new faces options with default values.
func FacesOptionsDefault(conf *config.Config) FacesOptions {
	result := FacesOptions{Threshold: conf.FaceClusterSample()}

	return result
}
