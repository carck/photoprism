package query

import (
	"fmt"
	"time"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/face"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/sanitize"
)

// Faces returns all (known / unmatched) faces from the index.
func Faces(knownOnly, unmatched, hidden bool) (result entity.Faces, err error) {
	stmt := Db()

	if unmatched {
		stmt = stmt.Where("matched_at IS NULL")
	}

	if knownOnly {
		stmt = stmt.Where("subj_uid <> ''")
	}

	err = stmt.Where("face_hidden = ?", hidden).
		Order("face_src DESC, subj_uid, samples DESC").
		Find(&result).Error

	return result, err
}

// ManuallyAddedFaces returns all manually added face clusters.
func ManuallyAddedFaces(hidden bool) (result entity.Faces, err error) {
	err = Db().
		Where("face_hidden = ?", hidden).
		Where("face_src = ?", entity.SrcManual).
		Where("subj_uid <> ''").Order("subj_uid, samples DESC").
		Find(&result).Error

	return result, err
}

func MatchFaces(force bool) (recoginized, unknowned int, err error) {
	sqlKnownFace := `
WITH ranked AS (
SELECT t.mid, t.face_id, t.subj_uid, t.dist,
		ROW_NUMBER() OVER (PARTITION BY t.mid ORDER BY t.dist ASC) AS rn
FROM (
	SELECT m.rowid AS mid,
			f.id AS face_id,
			f.subj_uid,
			f.sample_radius,
			distance_sqeuclidean_f32(m.embeddings_json, f.embedding_json) AS dist
	FROM markers m
	CROSS JOIN faces f
	WHERE m.subj_uid = ''
		AND f.subj_uid <> ''
		AND (m.matched_at IS NULL OR m.matched_at < f.created_at)
) AS t
WHERE t.dist < power(t.sample_radius + ?, 2)
)
UPDATE markers
SET subj_uid = ranked.subj_uid,
    face_id = ranked.face_id,
    face_dist = sqrt(ranked.dist),
	marker_review = 0,
    matched_at =  CURRENT_TIMESTAMP
FROM ranked
WHERE markers.rowid = ranked.mid
  AND ranked.rn = 1
RETURNING markers.marker_uid;`

	sqlUnknownFace := `
WITH ranked AS (
    SELECT t.mid, t.face_id, t.subj_uid, t.dist,
           ROW_NUMBER() OVER (PARTITION BY t.mid ORDER BY t.dist ASC) AS rn
    FROM (
        SELECT m.rowid AS mid,
               f.id AS face_id,
               f.subj_uid,
               f.sample_radius,
               distance_sqeuclidean_f32(m.embeddings_json, f.embedding_json) AS dist
        FROM markers m
        CROSS JOIN faces f
        WHERE m.face_id = ''
          AND m.subj_uid = ''
          AND f.subj_uid = ''
		  AND (m.matched_at IS NULL OR m.matched_at < f.created_at)
    ) AS t
    WHERE t.dist < power(t.sample_radius + ?, 2)
)
UPDATE markers
SET subj_uid = ranked.subj_uid,
    face_id = ranked.face_id,
    face_dist = sqrt(ranked.dist),
	marker_review = 0,
    matched_at =  CURRENT_TIMESTAMP
FROM ranked
WHERE markers.rowid = ranked.mid
  AND ranked.rn = 1
RETURNING markers.marker_uid;
`

	sqlUpdateMatchedAt := `
UPDATE markers
SET matched_at = CURRENT_TIMESTAMP
WHERE markers.marker_type = 'face' 
AND markers.face_id = ''
  AND (
      markers.matched_at IS NULL
	  or markers.matched_at < ?
  )
RETURNING rowid;
`
	var known []string
	var unknown []string
	var matched_ats []int64

	if force {
		if res := Db().Exec(`UPDATE markers SET matched_at = NULL, subj_uid = '', face_id = '' WHERE marker_type = ?`, entity.MarkerFace); res.Error != nil {
			log.Errorf("faces: %s (reset matched_at)", res.Error)
			return 0, 0, res.Error
		}
	}
	if res := Db().Raw(sqlKnownFace, face.MatchDist).Scan(&known); res.Error != nil {
		return 0, 0, res.Error
	}

	if res := Db().Raw(sqlUnknownFace, face.MatchDist).Scan(&unknown); res.Error != nil {
		return 0, 0, res.Error
	}

	var maxCreateAt time.Time
	if err := Db().Model(&entity.Face{}).Select("MAX(created_at)").Scan(&maxCreateAt).Error; err != nil {
		return 0, 0, err
	}

	if res := Db().Raw(sqlUpdateMatchedAt, maxCreateAt).Scan(&matched_ats); res.Error != nil {
		return 0, 0, res.Error
	}
	log.Infof("faces: matched %d known and %d unknown faces, updated %d matched_at timestamps", len(known), len(unknown), len(matched_ats))

	return len(known), len(unknown), nil
}

// MatchFaceMarkers matches markers with known faces.
func MatchFaceMarkers() (affected int64, err error) {
	faces, err := Faces(true, false, false)

	if err != nil {
		return affected, err
	}

	for _, f := range faces {
		if res := Db().Exec(`update markers set subj_uid=?, marker_review=0 
					where marker_invalid=0 and face_id=? and subj_src=? and subj_uid<>?`,
			f.SubjUID, f.ID, entity.SrcAuto, f.SubjUID); res.Error != nil {
			return affected, err
		} else if res.RowsAffected > 0 {
			affected += res.RowsAffected
		}
	}

	return affected, nil
}

// RemoveAnonymousFaceClusters removes anonymous faces from the index.
func RemoveAnonymousFaceClusters() (removed int64, err error) {
	res := UnscopedDb().
		Delete(entity.Face{}, "subj_uid = '' AND face_src = ?", entity.SrcAuto)

	return res.RowsAffected, res.Error
}

// RemoveAutoFaceClusters removes automatically added face clusters from the index.
func RemoveAutoFaceClusters() (removed int64, err error) {
	res := UnscopedDb().
		Delete(entity.Face{}, "face_src = ?", entity.SrcAuto)

	return res.RowsAffected, res.Error
}

func ShouldRunFaceMatch(since time.Time) bool {
	n := 0
	q := Db().Model(&entity.Markers{}).
		Where("subj_uid = ''").
		Where("created_at >= ?", since)

	if err := q.Count(&n).Error; err != nil || n > 0 {
		if err != nil {
			log.Errorf("faces: %s (face match)", err)
		}
		return true
	}

	q = Db().Model(&entity.Faces{}).
		Where("created_at >= ?", since)

	if err := q.Count(&n).Error; err != nil || n > 0 {
		if err != nil {
			log.Errorf("faces: %s (face match)", err)
		}
		return true
	}
	return false
}

// CountNewFaceMarkers counts the number of new face markers in the index.
func CountNewFaceMarkers(size, score int, since time.Time) (n int) {

	q := Db().Model(&entity.Markers{}).
		Where("marker_type = ?", entity.MarkerFace).
		Where("face_id = '' AND marker_invalid = 0 AND embeddings_json <> ''")

	if size > 0 {
		q = q.Where("size >= ?", size)
	}

	if score > 0 {
		q = q.Where("score >= ?", score)
	}

	if !since.IsZero() {
		q = q.Where("created_at > ?", since)
	}

	if err := q.Count(&n).Error; err != nil {
		log.Errorf("faces: %s (count new markers)", err)
	}

	return n
}

// PurgeOrphanFaces removes unused faces from the index.
func PurgeOrphanFaces(faceIds []string) (removed int64, err error) {
	// Remove invalid face IDs.
	if res := Db().
		Where("id IN (?)", faceIds).
		Where(fmt.Sprintf("id NOT IN (SELECT face_id FROM %s)", entity.Marker{}.TableName())).
		Delete(&entity.Face{}); res.Error != nil {
		return removed, fmt.Errorf("faces: %s while purging orphans", res.Error)
	} else {
		removed += res.RowsAffected
	}

	return removed, nil
}

// MergeFaces returns a new face that replaces multiple others.
func MergeFaces(merge entity.Faces) (merged *entity.Face, err error) {
	if len(merge) < 2 {
		// Nothing to merge.
		return merged, fmt.Errorf("faces: two or more clusters required for merging")
	}

	subjUID := merge[0].SubjUID

	for i := 1; i < len(merge); i++ {
		if merge[i].SubjUID != subjUID {
			return merged, fmt.Errorf("faces: cannot merge clusters with conflicting subjects %s <> %s",
				sanitize.Log(subjUID), sanitize.Log(merge[i].SubjUID))
		}
	}

	// Find or create merged face cluster.
	if merged = entity.NewFace(merge[0].SubjUID, merge[0].FaceSrc, merge.Embeddings()); merged == nil {
		return merged, fmt.Errorf("faces: new cluster is nil for subject %s", sanitize.Log(subjUID))
	} else if merged = entity.FirstOrCreateFace(merged); merged == nil {
		return merged, fmt.Errorf("faces: failed creating new cluster for subject %s", sanitize.Log(subjUID))
	} else if err := merged.MatchMarkers(append(merge.IDs(), "")); err != nil {
		return merged, err
	}

	// PurgeOrphanFaces removes unused faces from the index.
	if removed, err := PurgeOrphanFaces(merge.IDs()); err != nil {
		return merged, err
	} else if removed > 0 {
		log.Debugf("faces: removed %d orphans for subject %s", removed, sanitize.Log(subjUID))
	} else {
		log.Warnf("faces: failed removing merged clusters for subject %s", sanitize.Log(subjUID))
	}

	return merged, err
}

// ResolveFaceCollisions resolves collisions of different subject's faces.
func ResolveFaceCollisions() (conflicts, resolved int, err error) {
	faces, err := Faces(true, false, false)

	if err != nil {
		return conflicts, resolved, err
	}

	for _, f1 := range faces {
		for _, f2 := range faces {
			if matched, dist := f1.Match(face.Embeddings{f2.Embedding()}); matched {
				if f1.SubjUID == f2.SubjUID {
					continue
				}

				conflicts++

				r := f1.SampleRadius + face.MatchDist

				log.Infof("face %s: ambiguous subject at dist %f, Ø %f from %d samples, collision Ø %f", f1.ID, dist, r, f1.Samples, f1.CollisionRadius)

				if f1.SubjUID != "" {
					log.Debugf("face %s: subject %s (%s %s)", f1.ID, sanitize.Log(f1.SubjUID), f1.SubjUID, entity.SrcString(f1.FaceSrc))
				} else {
					log.Debugf("face %s: has no subject (%s)", f1.ID, entity.SrcString(f1.FaceSrc))
				}

				if f2.SubjUID != "" {
					log.Debugf("face %s: subject %s (%s %s)", f2.ID, sanitize.Log(f2.SubjUID), f2.SubjUID, entity.SrcString(f2.FaceSrc))
				} else {
					log.Debugf("face %s: has no subject (%s)", f2.ID, entity.SrcString(f2.FaceSrc))
				}

				if ok, err := f1.ResolveCollision(face.Embeddings{f2.Embedding()}); err != nil {
					log.Errorf("face %s: %s", f1.ID, err)
				} else if ok {
					log.Infof("face %s: conflict has been resolved", f1.ID)
					resolved++
				} else {
					log.Debugf("face %s: conflict could not be resolved", f1.ID)
				}
			}
		}
	}

	return conflicts, resolved, nil
}

// RemovePeopleAndFaces permanently removes all people, faces, and face markers.
func RemovePeopleAndFaces() (err error) {
	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	// Delete people.
	if err = UnscopedDb().Delete(entity.Subject{}, "subj_type = ?", entity.SubjPerson).Error; err != nil {
		return err
	}

	// Delete all faces.
	if err = UnscopedDb().Delete(entity.Face{}).Error; err != nil {
		return err
	}

	// Delete face markers.
	if err = UnscopedDb().Delete(entity.Marker{}, "marker_type = ?", entity.MarkerFace).Error; err != nil {
		return err
	}

	// Reset face counters.
	if err = Db().Exec(`update photots set photo_faces=0`).Error; err != nil {
		return err
	}

	// Reset people label.
	if label, err := LabelBySlug("people"); err != nil {
		return err
	} else if err = UnscopedDb().
		Delete(entity.PhotoLabel{}, "label_id = ?", label.ID).Error; err != nil {
		return err
	} else if err = label.Update("PhotoCount", 0); err != nil {
		return err
	}

	// Reset portrait label.
	if label, err := LabelBySlug("portrait"); err != nil {
		return err
	} else if err = UnscopedDb().
		Delete(entity.PhotoLabel{}, "label_id = ?", label.ID).Error; err != nil {
		return err
	} else if err = label.Update("PhotoCount", 0); err != nil {
		return err
	}

	return nil
}
