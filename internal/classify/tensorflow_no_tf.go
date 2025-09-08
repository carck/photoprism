//go:build NOTENSORFLOW
// +build NOTENSORFLOW

package classify

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"runtime/debug"
	"sync"
)

// TensorFlow is now just a wrapper around the external HTTP service.
type TensorFlow struct {
	mu         sync.Mutex
	disabled   bool
	serviceURL string
}

// New returns new TensorFlow instance configured for external service.
func New(_ string, disabled bool) *TensorFlow {
	return &TensorFlow{
		disabled:   disabled,
		serviceURL: "http://127.0.0.1:9001/classify?file=",
	}
}

func (t *TensorFlow) Init() (err error) {
	return nil
}

// File asks the server to classify the file.
func (t *TensorFlow) File(filename string) (result Labels, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.disabled {
		return result, nil
	}

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("classify: %s (panic)\nstack: %s", r, debug.Stack())
		}
	}()

	resp, err := http.Get(t.serviceURL + url.QueryEscape(filename))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("classify: server error %s", resp.Status)
	}

	// Server already returns exactly the Labels JSON
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}
