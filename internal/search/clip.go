package search

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	gc "github.com/patrickmn/go-cache"
)

type clipResponse struct {
	ImageIDs []uint `json:"image_ids"`
}

var clipCache = gc.New(300*time.Second, 60*time.Second)

func SearchClip(query string, size int) (photos []uint, err error) {
	if ids, found := clipCache.Get(query); found {
		return ids.([]uint), nil
	}
	resp, err := http.Get("http://127.0.0.1:9005/search_text?query=" + url.QueryEscape(query) + "&top_k=" + strconv.Itoa(size))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("clip: server error %s", resp.Status)
	}

	var result clipResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	clipCache.Set(query, result.ImageIDs, gc.DefaultExpiration)
	return result.ImageIDs, nil
}
