package main

import (
	"log"
	"os"

	"github.com/lord-server/lopater/pkg/world"
)

func main() {
	worldPath := os.Args[1]

	w, err := world.Open(worldPath)
	if err != nil {
		log.Fatalf("failed to open world at %v: %v", worldPath, err)
	}

	log.Printf("backend = %v", w.Metadata.BackendType)

	data, err := w.Storage.GetBlockData(world.Position{
		X: 0,
		Y: 0,
		Z: 0,
	})

	if err != nil {
		log.Fatal(err)
	} else {
		log.Fatal(data)
	}
}
