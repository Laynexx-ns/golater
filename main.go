package main

import (
	"log"

	golater "github.com/Laynexx-ns/golater/internal"
)

func main() {
	_, err := golater.Golater().Run()
	if err != nil {
		log.Fatalf("failed to start app | err: %v", err)
	}
}
