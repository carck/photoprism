package entity

import (
	"time"
)

type JobState struct {
	Name      string    `gorm:"primaryKey;type:varchar(50)"`
	Value     string    `gorm:"type:varchar(50)"`
	CreatedAt time.Time `json:"-" yaml:"-"`
	UpdatedAt time.Time `json:"-" yaml:"-"`
}

// TableName returns the entity database table name.
func (JobState) TableName() string {
	return "job_states"
}
