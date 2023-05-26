package meta

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/photoprism/photoprism/pkg/sanitize"
)

// JSON parses a json sidecar file (as used by Exiftool) and returns a Data struct.
func JSON(jsonName, originalName string) (data Data, err error) {
	err = data.JSON(jsonName, originalName)

	return data, err
}

// JSON parses a json sidecar file (as used by Exiftool) and returns a Data struct.
func (data *Data) JSON(jsonData, originalName string) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("metadata: %s in %s (json panic)\nstack: %s", e, sanitize.Log(jsonData), debug.Stack())
		}
	}()

	if data.All == nil {
		data.All = make(map[string]string)
	}

	if strings.Contains(jsonData, "ExifToolVersion") {
		return data.Exiftool([]byte(jsonData), originalName)
	} else if strings.Contains(jsonData, "albumData") {
		return data.GMeta([]byte(jsonData))
	} else if strings.Contains(jsonData, "photoTakenTime") {
		return data.GPhoto([]byte(jsonData))
	}

	log.Warnf("metadata: unknown json in %s", jsonData)

	return fmt.Errorf("unknown json in %s", jsonData)
}
