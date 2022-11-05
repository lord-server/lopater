package world

import (
	"github.com/lord-server/lopater/pkg/spatial"
	_ "github.com/mattn/go-sqlite3"
)

type Storage interface {
	GetBlockData(pos spatial.BlockPosition) ([]byte, error)
	SetBlockData(pos spatial.BlockPosition, data []byte) error
	Close()
}
