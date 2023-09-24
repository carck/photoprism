package search

// AlbumSlim represents an album search result.
type AlbumSlim struct {
	AlbumUID    string `json:"UID"`
	Thumb       string `json:"Thumb"`
	AlbumType   string `json:"Type"`
	AlbumTitle  string `json:"Title"`
	AlbumPath   string `json:"Path"`
	AlbumFilter string `json:"Filter"`
}

type AlbumResultsSlim []AlbumSlim
