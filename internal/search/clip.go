package search

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	gc "github.com/patrickmn/go-cache"
	"github.com/photoprism/photoprism/pkg/rnd"
)

type clipResponse struct {
	ImageIDs []uint `json:"image_ids"`
}

var clipCache = gc.New(300*time.Second, 60*time.Second)
var clipServer = "http://127.0.0.1:9005"
var clipHTTPGet = http.Get

func SearchClip(query string, size int) ([]uint, error) {
	cacheKey := fmt.Sprintf("%s:%d", query, size)

	if v, ok := clipCache.Get(cacheKey); ok {
		if ids, ok := v.([]uint); ok {
			return ids, nil
		}
	}

	params := url.Values{}
	params.Set("top_k", strconv.Itoa(size))

	if rnd.IsPPID(query, 'p') {
		id, err := lookupPhotoID(query)
		if err != nil {
			return nil, err
		}
		params.Set("id", strconv.FormatUint(uint64(id), 10))
	} else {
		params.Set("query", query)
	}

	resp, err := clipHTTPGet(clipServer + "/search_text?" + params.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("clip: server error %s", resp.Status)
	}

	var r clipResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}

	clipCache.Set(cacheKey, r.ImageIDs, gc.DefaultExpiration)
	return r.ImageIDs, nil
}

func lookupPhotoID(uid string) (uint, error) {
	var id uint
	err := UnscopedDb().
		Table("photos").
		Select("id").
		Where("photo_uid = ?", uid).
		Take(&id).Error
	return id, err
}
