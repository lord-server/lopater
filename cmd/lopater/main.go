package main

import (
	"log"
	"os"

	"github.com/lord-server/lopater/pkg/world"
)

func main() {
	worldPath := os.Args[1]

	world, err := world.Open(worldPath)
	if err != nil {
		log.Fatalf("failed to open world at %v: %v", worldPath, err)
	}

	log.Printf("backend = %v", world.Metadata.BackendType)
}
