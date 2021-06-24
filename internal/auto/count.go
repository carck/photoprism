package auto

import (
	"github.com/photoprism/photoprism/internal/entity"
)


func ResetCount() {
	entity.ResetCount()
}

func MustCount() bool {
	return entity.MustCount()
}

func DoCount() {
	err := entity.DoUpdatePhotoCounts()
	if err != nil {
		log.Errorf("auto-count: %s", err.Error())
	}
}
