package app

import (
	"github.com/sirupsen/logrus"
	"go_code_reviewer/pkg/log"
)

type ServiceInterface interface {
	Start()
	Close()
}

func RunService(service ServiceInterface) {
	log.Init(log.Config{
		Level:     logrus.InfoLevel,
		Env:       "development",
		LogToFile: true,
		FilePath:  "./service.log",
	})

	service.Start()
	service.Close()
}
