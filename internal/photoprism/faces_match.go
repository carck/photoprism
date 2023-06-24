package photoprism

import (
	"fmt"
	"sync"
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
	matchedAt := entity.TimePointer()

	if opt.Force || runMatch {
		faces, err := query.Faces(false, false, false)

		if err != nil || len(faces) == 0 {
			return result, err
		}

		if r, err := w.MatchFaces(faces, opt.Force, matchedAt); err != nil {
			return result, err
		} else {
			result.Add(r)
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
func (w *Faces) MatchFaces(faces entity.Faces, force bool, matchedBefore *time.Time) (result FacesMatchResult, err error) {
	matched := 0
	limit := 500
	max := query.CountMarkers(entity.MarkerFace)

	for {
		var markers entity.Markers

		if force {
			markers, err = query.FaceMarkers(limit, matched)
		} else {
			markers, err = query.UnmatchedFaceMarkers(limit, 0, matchedBefore)
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

			// Find the closest face match for marker.
			if !force && marker.MatchedAt != nil {
				skip := true
				for _, m := range faces {
					if m.CreatedAt.After(*marker.MatchedAt) {
						skip = false
					}
				}
				if skip {
					continue
				}
			}
			distResults := make([]FaceDistResult, len(faces))
			wg := new(sync.WaitGroup)
			wg.Add(len(faces))
			for i := range faces {
				go func(idx int) {
					f := &faces[idx]
					ok, dist := f.Match(marker.Embeddings())
					distResults[idx] = FaceDistResult{ok, dist}
				}(i)
			}
			wg.Wait()
			for i, r := range distResults {
				m := &faces[i]
				if ok, dist := r.Ok, r.Dist; ok && (f == nil || dist < d && f.FaceSrc == m.FaceSrc) {
					f = &faces[i]
					d = dist
				}
			}

			// Marker already has the best matching face?
			if !marker.HasFace(f, d) {
				// Marker needs a (new) face.
			} else {
				log.Debugf("faces: marker %s already has the best matching face %s with dist %f", marker.MarkerUID, marker.FaceID, marker.FaceDist)

				if err := marker.Matched(); err != nil {
					log.Warnf("faces: %s while updating marker %s match timestamp", err, marker.MarkerUID)
				}

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
