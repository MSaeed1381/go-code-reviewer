package log

import (
	"github.com/sirupsen/logrus"
	"io"
	"os"
)

var log *logrus.Logger

func Init(cfg Config) {
	log = logrus.New()
	log.SetLevel(cfg.Level)
	log.SetFormatter(&logrus.TextFormatter{ // development formatter
		FullTimestamp: true,
		ForceColors:   true,
	})
	if cfg.Env == "production" {
		log.SetFormatter(newJSONFormatter())
	}

	var writers []io.Writer
	writers = append(writers, os.Stdout)
	if cfg.LogToFile {
		file, err := os.OpenFile(cfg.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic(err)
		}
		writers = append(writers, file)
	}

	log.SetOutput(io.MultiWriter(writers...))
}

func GetLogger() *logrus.Logger {
	if log == nil {
		Init(Config{
			Level: logrus.InfoLevel,
			Env:   "development",
		})
	}
	return log
}
