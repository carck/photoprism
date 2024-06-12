package form

import "time"

// SearchPhotosSlim represents search form fields for "/api/v1/photos/slim".
type SearchPhotosSlim struct {
	Subject string    `form:"subject"`
	Album   string    `form:"album"`
	Path    string    `form:"path"`
	Notes   string    `fotm:"notes"`
	Before  time.Time `form:"before" time_format:"2006-01-02"` // Finds images taken before date
	Country string    `form:"country"`                         // Moments
	State   string    `form:"state"`                           // Moments
	Count   int       `form:"count" binding:"required" serialize:"-"`
	Offset  int       `form:"offset" serialize:"-"`
}
