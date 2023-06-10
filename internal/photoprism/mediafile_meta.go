package photoprism

import (
	"fmt"

	"github.com/photoprism/photoprism/internal/entity"

	"github.com/photoprism/photoprism/internal/meta"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/sanitize"
)

// HasSidecarJson returns true if this file has or is a json sidecar file.
func (m *MediaFile) HasSidecarJson() bool {
	if m.IsJson() {
		return true
	}

	return fs.FormatJson.FindFirst(m.FileName(), []string{Config().SidecarPath(), fs.HiddenPath}, Config().OriginalsPath(), false) != ""
}

// SidecarJsonName returns the corresponding JSON sidecar file name as used by Google Photos (and potentially other apps).
func (m *MediaFile) SidecarJsonName() string {
	jsonName := m.fileName + ".json"

	if fs.FileExists(jsonName) {
		return jsonName
	}

	return ""
}

// NeedsExifToolJson tests if an ExifTool JSON file needs to be created.
func (m *MediaFile) NeedsExifToolJson() bool {
	if m.Root() == entity.RootSidecar || !m.IsMedia() || m.IsSidecar() {
		return false
	}

	return true
}

// ReadExifToolJson reads metadata from a cached ExifTool JSON file.
func (m *MediaFile) ReadExifToolJson(jsonName string) error {
	return m.metaData.JSON(jsonName, "")
}

// MetaData returns exif meta data of a media file.
func (m *MediaFile) MetaData() (result meta.Data) {
	m.metaDataOnce.Do(func() {
		var err error

		if m.ExifSupported() {
			err = m.metaData.Exif(m.FileName(), m.FileType(), Config().ExifBruteForce())
		} else {
			err = fmt.Errorf("exif not supported")
		}

		if err != nil {
			//m.metaData.Error = err
			log.Debugf("metadata: %s in %s", err, sanitize.Log(m.BaseName()))
		}
	})

	return m.metaData
}
