package api

import (
	"fmt"
	"os"
)

func Etag(filename, match string) (etag string, changed bool) {
	fi, err := os.Stat(filename)
	if err != nil {
		return match, false
	}
	etag = fmt.Sprintf("%x-%x", fi.ModTime().Unix(), fi.Size())
	return etag, etag != match
}
