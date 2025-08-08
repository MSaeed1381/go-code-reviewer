package app

import (
	"github.com/sirupsen/logrus"
	"go_code_reviewer/pkg/log"
)

const (
	defaultLogFilePath = "../../service.log"
)

type ServiceInterface interface {
	Start()
	Close()
}

func RunService(serviceName string, service ServiceInterface) {
	log.Init(log.Config{
		Level:     logrus.InfoLevel,
		LogToFile: true,
		FilePath:  defaultLogFilePath,
		Service:   serviceName,
	})

	service.Start()
	service.Close()
}
