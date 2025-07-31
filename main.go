package main

import "go_code_reviewer/internal"

func main() {
	service := internal.Service{}
	service.Start()
	service.Close()
}
