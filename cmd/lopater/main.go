package main

import (
	"flag"
	"log"

	"github.com/lord-server/lopater/pkg/world"
)

var (
	ConfigPath = flag.String("config", "config.hjson", "Path to config")
)

func init() {
	flag.Parse()
}

func main() {
	if flag.NArg() < 1 {
		log.Fatalf("usage: lopater [-config /path/to/config.hjson] </path/to/world>")
	}
	worldPath := flag.Args()[0]

	w, err := world.Open(worldPath)
	if err != nil {
		log.Fatalf("failed to open world at %v: %v", worldPath, err)
	}

	log.Printf("backend = %v", w.Metadata.BackendType)

	block, err := w.GetBlock(world.Position{
		X: 0,
		Y: 0,
		Z: 0,
	})

	if err != nil {
		log.Fatal(err)
	} else {
		log.Fatal(block)
	}
}
