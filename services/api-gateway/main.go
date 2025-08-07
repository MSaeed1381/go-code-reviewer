package main

import (
	"go_code_reviewer/pkg/app"
	"go_code_reviewer/services/api-gateway/internal"
)

func main() {
	app.RunService("api-gateway", &internal.Service{})
}
