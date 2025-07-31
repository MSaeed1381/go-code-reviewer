package log

import (
	"github.com/sirupsen/logrus"
	"time"
)

func newJSONFormatter() logrus.Formatter {
	return &logrus.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "@timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
			logrus.FieldKeyFunc:  "caller",
		},
		PrettyPrint: false,
	}
}
