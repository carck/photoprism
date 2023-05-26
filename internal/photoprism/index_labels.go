package photoprism

import (
	"sort"
	"time"

	"github.com/photoprism/photoprism/internal/classify"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/sanitize"
)

// Labels classifies a JPEG image and returns matching labels.
func (ind *Index) Labels(jpeg *MediaFile) (results classify.Labels) {
	start := time.Now()

	sizes := []thumb.Name{thumb.Tile224}

	var labels classify.Labels

	for _, size := range sizes {
		filename, err := jpeg.Thumbnail(Config().ThumbPath(), size)

		if err != nil {
			log.Debugf("%s in %s", err, sanitize.Log(jpeg.BaseName()))
			continue
		}

		imageLabels, err := ind.tensorFlow.File(filename)

		if err != nil {
			log.Debugf("%s in %s", err, sanitize.Log(jpeg.BaseName()))
			continue
		}

		labels = append(labels, imageLabels...)
	}

	// Sort by priority and uncertainty
	sort.Sort(labels)

	var confidence int

	for _, label := range labels {
		if confidence == 0 {
			confidence = 100 - label.Uncertainty
		}

		if (100 - label.Uncertainty) > (confidence / 3) {
			results = append(results, label)
		}
	}

	if l := len(labels); l == 1 {
		log.Infof("index: matched %d label with %s [%s]", l, sanitize.Log(jpeg.BaseName()), time.Since(start))
	} else if l > 1 {
		log.Infof("index: matched %d labels with %s [%s]", l, sanitize.Log(jpeg.BaseName()), time.Since(start))
	}

	return results
}
