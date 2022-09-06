package world

import (
	"fmt"
	"log"
	"path/filepath"
)

type Position struct {
	X, Y, Z int32
}

type World struct {
	Metadata WorldMetadata
	Storage  Storage
}

// FIXME: this parser assumes that world.mt files are well-formed
// and not malicious
func Open(path string) (*World, error) {
	log.Printf("world path: %v", path)
	metadata, err := ReadMetadata(filepath.Join(path, "world.mt"))
	if err != nil {
		return nil, err
	}

	var storage Storage

	switch metadata.BackendType {
	case BackendSQLite:
		storage, err = openSQLite(filepath.Join(path, "map.sqlite"))
		if err != nil {
			return nil, fmt.Errorf("unable to open world: %w", err)
		}
	case BackendPostgreSQL:
		panic("unimplemented")
	}

	return &World{
		Metadata: metadata,
		Storage:  storage,
	}, nil
}
