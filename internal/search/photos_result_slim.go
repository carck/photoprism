package search

import (
	"time"
)

type PhotoResultsSlim []PhotoSlim

type PhotoSlim struct {
	ID         uint      `json:"-"`
	PhotoUID   string    `json:"UID"`
	PhotoType  string    `json:"Type"`
	TakenAt    time.Time `json:"TakenAt"`
	FileHash   string    `json:"Hash"`
	PhotoName  string    `json:"Name"`
	PhotoTitle string    `json:"Title"`
}

func (photos *PhotoResultsSlim) SortByIDs(ids []uint) {
	if len(ids) == 0 || len(*photos) == 0 {
		return
	}

	photoMap := make(map[uint]PhotoSlim, len(*photos))
	for _, p := range *photos {
		photoMap[p.ID] = p
	}

	sorted := make(PhotoResultsSlim, 0, len(*photos))
	for _, id := range ids {
		if p, ok := photoMap[id]; ok {
			sorted = append(sorted, p)
		}
	}

	*photos = sorted
}
