package photoprism

import (
	"fmt"
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/query"
)

var lastMatch time.Time = time.Unix(0, 0)

// FacesMatchResult represents the outcome of Faces.Match().
type FacesMatchResult struct {
	Updated    int64
	Recognized int64
	Unknown    int64
}

type FaceDistResult struct {
	Ok   bool
	Dist float64
}

// Add adds result counts.
func (r *FacesMatchResult) Add(result FacesMatchResult) {
	r.Updated += result.Updated
	r.Recognized += result.Recognized
	r.Unknown += result.Unknown
}

// Match matches markers with faces and subjects.
func (w *Faces) Match(opt FacesOptions) (result FacesMatchResult, err error) {
	if w.Disabled() {
		return result, fmt.Errorf("facial recognition is disabled")
	}

	var runMatch bool

	// Skip matching if index contains no new face markers, and force option isn't set.
	if opt.Force {
		log.Infof("faces: updating all markers")
	} else if runMatch = query.ShouldRunFaceMatch(lastMatch); runMatch {
		log.Infof("faces: run matches")
	} else {
		log.Debugf("faces: found no unmatched markers")
	}

	lastMatch = time.Now()

	if opt.Force || runMatch {
		if r, u, err := query.MatchFaces(opt.Force); err != nil {
			return result, err
		} else {
			result.Recognized += int64(r)
			result.Unknown += int64(u)
			log.Infof("faces: matched %s known and %s unknown faces", english.Plural(r, "face", "faces"), english.Plural(u, "face", "faces"))
		}
	}

	return result, nil
}
