package thumb

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func ResampleVips(source, fileName string, width, height int, opts ...ResampleOption) error {
	method, _, _ := ResampleOptions(opts...)

	// JPEG quality
	q := int(JpegQuality)
	if width <= 150 && height <= 150 {
		q = int(JpegQualitySmall)
	}

	// Build query parameters
	params := url.Values{}
	params.Set("source", source)
	params.Set("target", fileName)
	params.Set("w", fmt.Sprintf("%d", width))
	params.Set("h", fmt.Sprintf("%d", height))
	params.Set("q", fmt.Sprintf("%d", q))
	params.Set("method", string(method))

	// New HTTP service endpoint
	baseURL := "http://127.0.0.1:9000/resize" // change port if needed
	u := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// Make HTTP GET request
	resp, err := http.Get(u)
	if err != nil {
		return fmt.Errorf("failed to call vips service: %w", err)
	}
	defer resp.Body.Close()

	// Read response body (optional, for error messages)
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("vips service error: %s", string(body))
	}

	return nil
}
