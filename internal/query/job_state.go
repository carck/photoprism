package query

import (
	"github.com/photoprism/photoprism/internal/entity"
)

// GetJobState retrieves the value of a job state by its name.
func (q *Query) GetJobState(name string, defaultValue string) string {
	var state entity.JobState

	if err := q.db.Where("name = ?", name).First(&state).Error; err != nil {
		return defaultValue
	}

	return state.Value
}

// SetJobState sets or updates the value of a job state by its name.
func (q *Query) SetJobState(name, value string) error {
	var state JobState

	if err := q.db.Where("name = ?", name).First(&state).Error; err != nil {
		state.Name = name
		state.Value = value
		return q.db.Create(&state).Error
	}

	state.Value = value
	return q.db.Save(&state).Error
}
