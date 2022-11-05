package world

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/lord-server/lopater/pkg/block"
	"github.com/lord-server/lopater/pkg/spatial"
)

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

	var s Storage

	switch metadata.BackendType {
	case BackendSQLite:
		s, err = openSQLite(filepath.Join(path, "map.sqlite"))
	case BackendPostgreSQL:
		params, ok := metadata.Variables["pgsql_connection"]
		if !ok {
			params = ""
		}
		s, err = openPostgres(params)
	}
	if err != nil {
		return nil, fmt.Errorf("unable to open world: %w", err)
	}

	return &World{
		Metadata: metadata,
		Storage:  s,
	}, nil
}

func (w *World) GetBlock(pos spatial.BlockPosition) (*block.MapBlock, error) {
	data, err := w.Storage.GetBlockData(pos)
	if err != nil {
		return nil, fmt.Errorf("failed to %w", err)
	}

	// Block doesn't exist
	if data == nil {
		return nil, nil
	}

	mapBlock, err := block.Decode(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode MapBlock: %w", err)
	}

	return mapBlock, nil
}

func (w *World) Close() {
	w.Storage.Close()
}
