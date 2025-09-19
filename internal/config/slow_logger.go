package config

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// SlowLogger is a custom GORM logger that logs slow SQL queries.
type SlowLogger struct {
	threshold time.Duration
	logger    *logrus.Logger
}

func (l *SlowLogger) Print(values ...interface{}) {
	if len(values) > 0 && values[0] == "sql" {
		if duration, ok := values[2].(time.Duration); ok && duration > l.threshold {
			l.logger.Print(fmt.Sprintln(values...))
		}
	} else {
		l.logger.Print(fmt.Sprintln(values...))
	}
}
