package photoprism

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/query"
)

// Faces represents a worker for face clustering and matching.
type Faces struct {
	conf *config.Config
}

// NewFaces returns a new Faces worker.
func NewFaces(conf *config.Config) *Faces {
	instance := &Faces{
		conf: conf,
	}

	return instance
}

// StartDefault starts face clustering and matching with default options.
func (w *Faces) StartDefault() (err error) {
	return w.Start(FacesOptions{
		Force:     false,
		Threshold: conf.FaceClusterSample(),
	})
}

func (w *Faces) DoMatch(opt FacesOptions) {
	var start time.Time

	start = time.Now()
	matches, err := w.Match(opt)

	if err != nil {
		log.Errorf("faces: %s (match)", err)
	}

	// Log face matching results.
	if matches.Updated > 0 {
		log.Infof("faces: updated %s, recognized %s, %d unknown [%s]", english.Plural(int(matches.Updated), "marker", "markers"), english.Plural(int(matches.Recognized), "face", "faces"), matches.Unknown, time.Since(start))
	} else {
		log.Debugf("faces: updated %s, recognized %s, %d unknown [%s]", english.Plural(int(matches.Updated), "marker", "markers"), english.Plural(int(matches.Recognized), "face", "faces"), matches.Unknown, time.Since(start))
	}
}

// Start face clustering and matching.
func (w *Faces) Start(opt FacesOptions) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s (panic)\nstack: %s", r, debug.Stack())
			log.Errorf("faces: %s", err)
		}
	}()

	if w.Disabled() {
		return fmt.Errorf("facial recognition is disabled")
	}

	if err := mutex.FacesWorker.Start(); err != nil {
		return err
	}

	defer mutex.FacesWorker.Stop()

	var start time.Time

	// Remove orphan file markers.
	start = time.Now()
	if removed, err := query.RemoveOrphanMarkers(); err != nil {
		log.Errorf("faces: %s (remove orphan markers)", err)
	} else if removed > 0 {
		log.Infof("faces: removed %d orphan markers [%s]", removed, time.Since(start))
	} else {
		log.Debugf("faces: found no orphan markers [%s]", time.Since(start))
	}

	// Repair invalid marker face and subject references.
	start = time.Now()
	if removed, err := query.FixMarkerReferences(); err != nil {
		log.Errorf("markers: %s (fix references)", err)
	} else if removed > 0 {
		log.Infof("markers: fixed %d references [%s]", removed, time.Since(start))
	} else {
		log.Debugf("markers: found no invalid references [%s]", time.Since(start))
	}

	// Create known marker subjects if needed.
	start = time.Now()
	if affected, err := query.CreateMarkerSubjects(); err != nil {
		log.Errorf("markers: %s (create subjects)", err)
	} else if affected > 0 {
		log.Infof("markers: added %d known subjects [%s]", affected, time.Since(start))
	} else {
		log.Debugf("markers: found no missing subjects [%s]", time.Since(start))
	}

	// Optimize existing face clusters.
	start = time.Now()
	if res, err := w.Optimize(); err != nil {
		return err
	} else if res.Merged > 0 {
		log.Infof("faces: merged %d clusters [%s]", res.Merged, time.Since(start))
	} else {
		log.Debugf("faces: found no clusters to be merged [%s]", time.Since(start))
	}

	w.DoMatch(opt)

	var added entity.Faces

	// Cluster existing face embeddings.
	start = time.Now()
	if added, err = w.Cluster(opt); err != nil {
		log.Errorf("faces: %s (cluster)", err)
	} else if n := len(added); n > 0 {
		log.Infof("faces: added %d new faces [%s]", n, time.Since(start))
	} else {
		log.Debugf("faces: found no new faces [%s]", time.Since(start))
	}

	// Match markers with faces and subjects.
	w.DoMatch(opt)

	return nil
}

// Cancel stops the current operation.
func (w *Faces) Cancel() {
	mutex.FacesWorker.Cancel()
}

// Canceled tests if face clustering and matching should be stopped.
func (w *Faces) Canceled() bool {
	return mutex.FacesWorker.Canceled() || mutex.MainWorker.Canceled() || mutex.MetaWorker.Canceled()
}

// Disabled tests if facial recognition is disabled.
func (w *Faces) Disabled() bool {
	return w.conf.DisableFaces()
}
