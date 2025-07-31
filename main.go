package main

import (
	"go_code_reviewer/internal"
	"go_code_reviewer/pkg/app"
)

func main() {
	app.RunService(&internal.Service{})
}
