package viewer

import (
	"time"
)

// Thumbs represents photo viewer thumbs in different sizes.
type Thumbs struct {
	Fit1280 Thumb `json:"fit_1280"`
	Fit1920 Thumb `json:"fit_1920"`
}

// Result represents a photo viewer result.
type Result struct {
	UID          string    `json:"UID"`
	Title        string    `json:"Title"`
	TakenAtLocal time.Time `json:"TakenAtLocal"`
	Description  string    `json:"Description"`
	Favorite     bool      `json:"Favorite"`
	Playable     bool      `json:"Playable"`
	DownloadUrl  string    `json:"DownloadUrl"`
	Width        int       `json:"Width"`
	Height       int       `json:"Height"`
	Thumbs       Thumbs    `json:"Thumbs"`
}

// Results represents a list of viewer search results.
type Results []Result
