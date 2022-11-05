package main

import (
	"flag"
	"log"

	"github.com/lord-server/lopater/pkg/spatial"
	"github.com/lord-server/lopater/pkg/world"
)

func init() {
	flag.Parse()
}

func main() {
	if flag.NArg() < 1 {
		log.Fatalf("usage: lopater </path/to/world>")
	}
	worldPath := flag.Args()[0]

	w, err := world.Open(worldPath)
	if err != nil {
		log.Fatalf("failed to open world at %v: %v", worldPath, err)
	}
	defer w.Close()

	log.Printf("backend = %v", w.Metadata.BackendType)

	mapBlock, err := w.GetMapBlock(spatial.MapBlockPosition{
		X: 0,
		Y: 0,
		Z: 0,
	})

	if err != nil {
		log.Fatal(err)
	} else {
		log.Fatal(mapBlock)
	}

}
