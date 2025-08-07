package app

import (
	"github.com/sirupsen/logrus"
	"go_code_reviewer/pkg/log"
)

type ServiceInterface interface {
	Start()
	Close()
}

func RunService(serviceName string, service ServiceInterface) {
	log.Init(log.Config{
		Level:     logrus.InfoLevel,
		Env:       "development",
		LogToFile: true,
		FilePath:  "../../service.log",
		Service:   serviceName,
	})

	service.Start()
	service.Close()
}
