package log

import (
	"github.com/sirupsen/logrus"
	"io"
	"log"
	"os"
)

type Field string

const (
	ServiceName Field = "service_name"
)

var logEntry *logrus.Entry

func Init(cfg Config) {
	logger := logrus.New()
	logger.SetLevel(cfg.Level)
	logger.SetFormatter(newJSONFormatter())

	var writers []io.Writer
	writers = append(writers, os.Stdout)
	if cfg.LogToFile {
		file, err := os.OpenFile(cfg.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic(err)
		}
		writers = append(writers, file)
	}

	logEntry = logger.WithFields(logrus.Fields{string(ServiceName): cfg.Service})
	log.SetOutput(io.MultiWriter(writers...))
}

func GetLogger() *logrus.Entry {
	if logEntry == nil {
		Init(Config{
			Level: logrus.InfoLevel,
			Env:   "development",
		})
	}
	return logEntry
}
