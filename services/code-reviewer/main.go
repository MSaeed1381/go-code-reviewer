package main

import (
	"go_code_reviewer/pkg/app"
	"go_code_reviewer/services/code-reviewer/internal"
)

func main() {
	app.RunService(&internal.Service{})
}
