package log

import "github.com/sirupsen/logrus"

type Config struct {
	Level     logrus.Level
	Env       string // "production" or "development"
	LogToFile bool
	FilePath  string
}
