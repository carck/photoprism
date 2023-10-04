package photoprism

import (
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/sanitize"
)

// Ocr return text ocr result
func (ind *Index) Ocr(jpeg *MediaFile) string {
	start := time.Now()

	size := thumb.Fit1280

	thumbName, err := jpeg.Thumbnail(Config().ThumbPath(), size)

	if err != nil {
		log.Debugf("index: %s in %s (ocr)", err, sanitize.Log(jpeg.BaseName()))
		return ""
	}

	resp, err := http.Get("http://localhost:8009?f=" + url.QueryEscape(thumbName))

	if err != nil {
		return ""
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	log.Infof("index: ocr for %s [%s]", sanitize.Log(jpeg.BaseName()), time.Since(start))

	return string(body)
}
