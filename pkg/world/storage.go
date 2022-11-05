package world

import (
	"github.com/lord-server/lopater/pkg/spatial"
)

// MapStorage abstracts map storage facilities and acts as a key-value database.
// As of Minetest 5.7, each coordinate uses only 12 bits (and packing position
// inside 64-bit integer is possible), although implementors shouldn't rely on
// this fact.
type MapStorage interface {
	GetMapBlockData(pos spatial.MapBlockPosition) ([]byte, error)
	SetMapBlockData(pos spatial.MapBlockPosition, data []byte) error
	Close()
}
