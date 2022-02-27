package form

// SearchPhotosSlim represents search form fields for "/api/v1/photos/slim".
type SearchPhotosSlim struct {
	Subject string `form:"subject"`
	Album   string `form:"album"`
	Path    string `form:"path"`
	Count   int    `form:"count" binding:"required" serialize:"-"`
	Offset  int    `form:"offset" serialize:"-"`
}
