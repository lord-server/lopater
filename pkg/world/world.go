package world

import (
	"fmt"
	"path/filepath"

	"github.com/lord-server/lopater/pkg/mapblock"
	"github.com/lord-server/lopater/pkg/spatial"
)

type World struct {
	Metadata WorldMetadata
	Storage  MapStorage
}

func Open(path string) (*World, error) {
	metadata, err := ReadMetadata(filepath.Join(path, "world.mt"))
	if err != nil {
		return nil, err
	}

	var storage MapStorage

	switch metadata.BackendType {
	case BackendSQLite:
		storage, err = openSQLite(filepath.Join(path, "map.sqlite"))
	case BackendPostgreSQL:
		params, ok := metadata.Variables["pgsql_connection"]
		if !ok {
			params = ""
		}
		storage, err = openPostgres(params)
	}

	if err != nil {
		return nil, fmt.Errorf("unable to open world: %w", err)
	}

	return &World{
		Metadata: metadata,
		Storage:  storage,
	}, nil
}

func (w *World) GetMapBlock(pos spatial.MapBlockPosition) (*mapblock.MapBlock, error) {
	data, err := w.Storage.GetMapBlockData(pos)
	if err != nil {
		return nil, fmt.Errorf("failed to %w", err)
	}

	// MapBlock doesn't exist
	if data == nil {
		return nil, nil
	}

	mapBlock, err := mapblock.Decode(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode MapBlock: %w", err)
	}

	return mapBlock, nil
}

func (w *World) Close() {
	w.Storage.Close()
}
