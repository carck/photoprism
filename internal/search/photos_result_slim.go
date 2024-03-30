package search

import (
	"time"
)

type PhotoResultsSlim []PhotoSlim

type PhotoSlim struct {
	PhotoUID   string    `json:"UID"`
	PhotoType  string    `json:"Type"`
	TakenAt    time.Time `json:"TakenAt"`
	FileHash   string    `json:"Hash"`
	PhotoName  string    `json:"Name"`
	PhotoTitle string    `json:"Title"`
}
