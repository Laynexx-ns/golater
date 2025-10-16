package main

import (
	golater "golater/internal"
	"log"
)

func main() {
	_, err := golater.Golater().Run()
	if err != nil {
		log.Fatalf("failed to start app | err: %v", err)
	}
}
