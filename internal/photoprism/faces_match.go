package photoprism

import (
	"fmt"
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/query"
)

var lastMatch time.Time = time.Unix(0, 0)

// FacesMatchResult represents the outcome of Faces.Match().
type FacesMatchResult struct {
	Updated    int64
	Recognized int64
	Unknown    int64
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
	matchedAt := entity.TimePointer()

	if opt.Force || runMatch {
		faces, err := query.Faces(false, false, false)

		if err != nil || len(faces) == 0 {
			return result, err
		}

		if r, err := w.MatchFaces(faces, opt.Force); err != nil {
			return result, err
		} else {
			result.Add(r)
		}
		for _, m := range faces {
			if err := m.Matched(matchedAt); err != nil {
				log.Warnf("faces: %s (update match timestamp)", err)
			}
		}
	}

	// Find unmatched faces.
	if unmatchedFaces, err := query.Faces(false, true, false); err != nil {
		log.Error(err)
	} else if len(unmatchedFaces) > 0 {
		if r, err := w.MatchFaces(unmatchedFaces, false); err != nil {
			return result, err
		} else {
			result.Add(r)
		}

		for _, m := range unmatchedFaces {
			if err := m.Matched(matchedAt); err != nil {
				log.Warnf("faces: %s (update match timestamp)", err)
			}
		}
	}

	// Update remaining markers based on previous matches.
	if m, err := query.MatchFaceMarkers(); err != nil {
		return result, err
	} else {
		result.Recognized += m
	}

	return result, nil
}

// MatchFaces matches markers against a slice of faces.
func (w *Faces) MatchFaces(faces entity.Faces, force bool) (result FacesMatchResult, err error) {
	matched := 0
	limit := 500
	max := query.CountMarkers(entity.MarkerFace)

	for {
		var markers entity.Markers

		if force {
			markers, err = query.FaceMarkers(limit, matched)
		} else {
			markers, err = query.UnmatchedFaceMarkers(limit, matched)
		}

		if err != nil {
			return result, err
		}
		if len(markers) == 0 {
			break
		}

		for _, marker := range markers {
			matched++

			if w.Canceled() {
				return result, fmt.Errorf("worker canceled")
			}

			// Skip invalid markers.
			if marker.MarkerInvalid || marker.MarkerType != entity.MarkerFace || len(marker.EmbeddingsJSON) == 0 {
				continue
			}

			// Pointer to the matching face.
			var f *entity.Face

			// Distance to the matching face.
			var d float64

			execute := false

			for _, m := range faces {
				if m.MatchedAt != nil && m.MatchedAt.After(marker.CreatedAt) {
					continue
				}
				execute = true
			}
			if !execute {
				continue
			}

			// Find the closest face match for marker.
			for i, m := range faces {
				if ok, dist := m.Match(marker.Embeddings()); ok && (f == nil || dist < d && f.FaceSrc == m.FaceSrc) {
					f = &faces[i]
					d = dist
				}
			}

			// Marker already has the best matching face?
			if !marker.HasFace(f, d) {
				// Marker needs a (new) face.
			} else {
				log.Debugf("faces: marker %s already has the best matching face %s with dist %f", marker.MarkerUID, marker.FaceID, marker.FaceDist)

				continue
			}

			// No matching face?
			if f == nil {
				if updated, err := marker.ClearFace(); err != nil {
					log.Warnf("faces: %s (clear marker face)", err)
				} else if updated {
					result.Updated++
				}

				continue
			}

			// Assign matching face to marker.
			updated, err := marker.SetFace(f, d)

			if err != nil {
				log.Warnf("faces: %s while setting a face for marker %s", err, marker.MarkerUID)
				continue
			}

			if updated {
				result.Updated++
			}

			if marker.SubjUID != "" {
				result.Recognized++
			} else {
				result.Unknown++
			}
		}

		log.Debugf("faces: matched %s", english.Plural(matched, "marker", "markers"))

		if matched > max {
			break
		}

		time.Sleep(50 * time.Millisecond)
	}

	return result, err
}
