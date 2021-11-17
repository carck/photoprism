package face

import (
	"fmt"

	"github.com/photoprism/photoprism/internal/crop"
)

// Areas is a list of face landmark areas.
type Areas []Area

// Relative returns all areas with relative coordinates.
func (pts Areas) Relative(r Area, rows, cols float32) (result crop.Areas) {
	for _, p := range pts {
		result = append(result, p.Relative(r, rows, cols))
	}

	return result
}

// Area represents a face landmark position.
type Area struct {
	Name string `json:"name,omitempty"`
	Row  int    `json:"x,omitempty"`
	Col  int    `json:"y,omitempty"`
	W    int    `json:"size,omitempty"`
	H    int    `json:"size,omitempty"`
}

// String returns the face landmark position as string.
func (a Area) String() string {
	return fmt.Sprintf("%d-%d-%d-%d", a.Row, a.Col, a.W, a.H)
}

// NewArea returns new face landmark coordinates.
func NewArea(name string, row, col, w, h int) Area {
	return Area{
		Name: name,
		Row:  row,
		Col:  col,
		W:    w,
		H:    h,
	}
}

// Relative returns the area with relative coordinates.
func (a Area) Relative(r Area, rows, cols float32) crop.Area {
	if rows < 1 {
		rows = 1
	}

	if cols < 1 {
		cols = 1
	}

	return crop.NewArea(
		a.Name,
		float32(a.Col-r.Col)/cols,
		float32(a.Row-r.Row)/rows,
		float32(a.W)/cols,
		float32(a.H)/rows,
	)
}

// TopLeft returns the top left position of the area.
func (a Area) TopLeft() (int, int) {
	return a.Row - (a.H / 2), a.Col - (a.W / 2)
}
