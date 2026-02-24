package thumb

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

func ResampleVipsDefault(source, target, hash string, force bool) error {
	q := int(JpegQuality)
	params := url.Values{}
	params.Set("source", source)
	params.Set("target", target)
	params.Set("hash", hash)
	params.Set("max", strconv.Itoa(SizePrecached))
	params.Set("q", strconv.Itoa(q))
	params.Set("force", strconv.FormatBool(force))

	baseURL := "http://127.0.0.1:9000/resize_default"
	u := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := http.Get(u)
	if err != nil {
		return fmt.Errorf("failed to call vips service: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("vips service error: %s", string(body))
	}

	return nil
}

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
	params.Set("w", strconv.Itoa(width))
	params.Set("h", strconv.Itoa(height))
	params.Set("q", strconv.Itoa(q))
	params.Set("method", string(method))

	baseURL := "http://127.0.0.1:9000/resize"
	u := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := http.Get(u)
	if err != nil {
		return fmt.Errorf("failed to call vips service: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("vips service error: %s", string(body))
	}

	return nil
}
