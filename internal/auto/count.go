package auto

import (
	"github.com/photoprism/photoprism/internal/entity"
)

func DoPhotoCount() {
	err := entity.DoUpdateCounts()
	if err != nil {
		log.Errorf("auto-count: %s", err.Error())
	}
}

func DoRefreshPhotos() {
	err := entity.DoRefreshPhotos()
	if err != nil {
		log.Errorf("auto refresh phot0: %s", err.Error())
	}
}
