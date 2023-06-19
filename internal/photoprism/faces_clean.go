package photoprism

import (
	"time"

	"github.com/photoprism/photoprism/internal/entity"
)

func (w *Faces) Clean() (err error) {
	// Remove unused people.
	start := time.Now()
	if count, err := entity.DeleteOrphanPeople(); err != nil {
		log.Errorf("faces: %s (remove people)", err)
	} else if count > 0 {
		log.Debugf("faces: removed %d people [%s]", count, time.Since(start))
	}

	// Remove unused face clusters.
	start = time.Now()
	if count, err := entity.DeleteOrphanFaces(); err != nil {
		log.Errorf("faces: %s (remove clusters)", err)
	} else if count > 0 {
		log.Debugf("faces: removed %d clusters [%s]", count, time.Since(start))
	}

	return nil
}
