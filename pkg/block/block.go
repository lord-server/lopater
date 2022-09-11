package block

import (
	"fmt"
	"io"

	"github.com/klauspost/compress/zstd"
)

// MinSupportedVersion and MaxSupportedVersion provide version range
// where correct encoding is guaranteed to be correct. Using decoder on MapBlocks
// outside of this range will produce an error.
const (
	MinSupportedVersion = 25
	MaxSupportedVersion = 29
)

// BlockSize defines the side length of a MapBlock in nodes
const BlockSize = 16

// BlockVolume defines the volume of a MapBlock in nodes
const BlockVolume = BlockSize * BlockSize * BlockSize

// NodeSizeInBytes is the amount of space (in bytes) required to store a single
// node in a non-compressed MapBlock
const NodeSizeInBytes = 4

type Node struct {
	ID     uint16
	Param1 uint8
	Param2 uint8
}

type MapBlock struct {
	Flags            uint8
	LightingComplete uint16
	Timestamp        uint32

	Mappings map[uint16]string
	NodeData []byte
	NodeMeta []byte
}

func readMappings(reader *binaryReader) (map[uint16]string, error) {
	mappingCount, err := reader.ReadUint16()
	if err != nil {
		return nil, err
	}

	mappings := make(map[uint16]string)
	for i := 0; i < int(mappingCount); i++ {
		id, err := reader.ReadUint16()
		if err != nil {
			return nil, err
		}
		name, err := reader.ReadString()
		if err != nil {
			return nil, err
		}

		mappings[id] = name
	}

	return mappings, nil
}

func decodeLegacyBlock(reader *binaryReader, version uint8) (*MapBlock, error) {
	var mapBlock MapBlock
	var err error

	mapBlock.Flags, err = reader.ReadUint8()
	if err != nil {
		return nil, fmt.Errorf("unable to decode flags: %w", err)
	}

	if version >= 27 {
		mapBlock.LightingComplete, err = reader.ReadUint16()
		if err != nil {
			return nil, fmt.Errorf("unable to decode lighting flags: %w", err)
		}
	}

	// Skip constant values:
	// - uint8 content_width
	// - uint8 params_width
	_, err = reader.Seek(1+1, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	mapBlock.NodeData, err = reader.ReadZlib()
	if err != nil {
		return nil, fmt.Errorf("unable to decode node data: %w", err)
	}

	mapBlock.NodeMeta, err = reader.ReadZlib()
	if err != nil {
		return nil, fmt.Errorf("unable to decode node metadata: %w", err)
	}

	// Skip constant value:
	// - uint8 staticObjectVersion
	_, err = reader.Seek(1, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	staticObjectCount, err := reader.ReadUint16()
	if err != nil {
		return nil, fmt.Errorf("unable to decode static object count: %w", err)
	}

	for i := 0; i < int(staticObjectCount); i++ {
		// - uint8 type
		// - int32 x, y, z
		_, err = reader.Seek(1+4+4+4, io.SeekCurrent)
		if err != nil {
			return nil, err
		}
		dataSize, err := reader.ReadUint16()
		if err != nil {
			return nil, err
		}
		_, err = reader.Seek(int64(dataSize), io.SeekCurrent)
		if err != nil {
			return nil, err
		}
	}

	// - uint32 timestamp
	mapBlock.Timestamp, err = reader.ReadUint32()
	if err != nil {
		return nil, fmt.Errorf("unable to decode timestamp: %w", err)
	}

	// Skip constant value:
	// - uint8 mappingVersion
	_, err = reader.Seek(1, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	mapBlock.Mappings, err = readMappings(reader)
	if err != nil {
		return nil, err
	}

	return &mapBlock, nil
}

func decodeBlock(reader *binaryReader) (*MapBlock, error) {
	z, err := zstd.NewReader(reader)
	if err != nil {
		return nil, err
	}
	defer z.Close()

	data, err := io.ReadAll(z)
	if err != nil {
		return nil, err
	}

	reader = newBinaryReader(data)

	var mapBlock MapBlock

	mapBlock.Flags, err = reader.ReadUint8()
	if err != nil {
		return nil, fmt.Errorf("unable to decode flags: %w", err)
	}

	mapBlock.LightingComplete, err = reader.ReadUint16()
	if err != nil {
		return nil, fmt.Errorf("unable to decode lighting flags: %w", err)
	}

	mapBlock.Timestamp, err = reader.ReadUint32()
	if err != nil {
		return nil, err
	}

	// Skip constant value:
	// - uint8 mapping version
	_, err = reader.Seek(1+2+4+1, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	mapBlock.Mappings, err = readMappings(reader)
	if err != nil {
		return nil, err
	}

	// Skip uint8 contentWidth, uint8 paramsWidth
	_, err = reader.Seek(1+1, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	nodeData := make([]byte, BlockVolume*NodeSizeInBytes)
	_, err = io.ReadFull(reader, nodeData)
	if err != nil {
		return nil, err
	}

	return &mapBlock, nil
}

func Decode(data []byte) (*MapBlock, error) {
	reader := newBinaryReader(data)

	version, err := reader.ReadUint8()
	if err != nil {
		return nil, err
	}

	if version < 29 {
		mapblock, err := decodeLegacyBlock(reader, version)
		if err != nil {
			return nil, err
		}
		return mapblock, nil
	}

	return decodeBlock(reader)
}
