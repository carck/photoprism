package photoprism

import (
	"strings"

	"github.com/photoprism/photoprism/pkg/sanitize"
)

type ConvertJob struct {
	file    *MediaFile
	convert *Convert
}

func ConvertWorker(jobs <-chan ConvertJob) {
	logError := func(err error, job ConvertJob) {
		fileName := job.file.RelName(job.convert.conf.OriginalsPath())
		log.Errorf("convert: %s for %s", strings.TrimSpace(err.Error()), sanitize.Log(fileName))
	}

	for job := range jobs {
		switch {
		case job.file == nil:
			continue
		case job.convert == nil:
			continue
		case job.file.IsVideo():
			if jsonName, err := job.convert.ToJson(job.file); err != nil {
				log.Debugf("convert: %s in %s (extract metadata)", sanitize.Log(err.Error()), sanitize.Log(job.file.BaseName()))
			} else if err := job.file.ReadExifToolJson(jsonName); err != nil {
				log.Errorf("convert: %s in %s (read metadata)", sanitize.Log(err.Error()), sanitize.Log(job.file.BaseName()))
			}

			if _, err := job.convert.ToJpeg(job.file); err != nil {
				logError(err, job)
			}
		default:
			if _, err := job.convert.ToJpeg(job.file); err != nil {
				logError(err, job)
			}
		}
	}
}
