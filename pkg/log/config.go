package log

import "github.com/sirupsen/logrus"

type Config struct {
	Level     logrus.Level
	LogToFile bool
	FilePath  string
	Service   string
}
