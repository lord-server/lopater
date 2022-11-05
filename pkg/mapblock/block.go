package mapblock

// MinSupportedVersion and MaxSupportedVersion provide version range where
// decoding is guaranteed to work. Using decoder on MapBlocks outside of
// this range will produce an error.
const (
	MinSupportedVersion = 25
	MaxSupportedVersion = 29
)

// MapBlockSize defines the side length of a single MapBlock in nodes
const MapBlockSize = 16

// MapBlockVolume defines the volume of a single MapBlock in nodes
const MapBlockVolume = MapBlockSize * MapBlockSize * MapBlockSize

// NodeSizeInBytes is the amount of space (in bytes) required to store a single
// node in a non-compressed MapBlock
const NodeSizeInBytes = 4

// Node represents a node (usually a cube) inside MapBlock
type Node struct {
	ID     uint16
	Param1 uint8
	Param2 uint8
}

type StaticObject struct {
	Type    uint8
	X, Y, Z int32
	Data    []byte
}

type NodeTimer struct {
	Position uint16
	Timeout  int32
	Elapsed  int32
}

type MapBlock struct {
	Flags            uint8
	LightingComplete uint16
	Timestamp        uint32

	Mappings map[uint16]string
	NodeData []byte
	NodeMeta []byte

	StaticObjects []StaticObject
	NodeTimers    []NodeTimer
}
