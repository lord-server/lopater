package block

import (
	"io"

	"github.com/klauspost/compress/zstd"
)

const BlockSize = 16
const BlockVolume = BlockSize * BlockSize * BlockSize
const NodeSizeInBytes = 4

type Node struct {
	ID     uint16
	Param1 uint8
	Param2 uint8
}

type MapBlock struct {
	Mappings map[uint16]string
	NodeData []byte
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
	if version >= 27 {
		// - uint8 flags
		// - uint16 lighting_complete
		// - uint8 content_width
		// - uint8 params_width
		_, err := reader.Seek(1+2+1+1, io.SeekCurrent)
		if err != nil {
			return nil, err
		}
	} else {
		// - uint8 flags
		// - uint8 content_width
		// - uint8 params_width
		_, err := reader.Seek(1+1+1, io.SeekCurrent)
		if err != nil {
			return nil, err
		}
	}

	nodeData, err := reader.ReadZlib()
	if err != nil {
		return nil, err
	}

	_, err = reader.ReadZlib()
	if err != nil {
		return nil, err
	}

	// - uint8 staticObjectVersion
	_, err = reader.Seek(1, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	staticObjectCount, err := reader.ReadUint16()
	if err != nil {
		return nil, err
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
	// - uint8 mappingVersion
	_, err = reader.Seek(4+1, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	mappings, err := readMappings(reader)
	if err != nil {
		return nil, err
	}

	return &MapBlock{
		Mappings: mappings,
		NodeData: nodeData,
	}, nil
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

	// Skip:
	// - uint8 flags
	// - uint16 lighting_complete
	// - uint32 timestamp
	// - uint8 mapping version
	_, err = reader.Seek(1+2+4+1, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	mappings, err := readMappings(reader)
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

	return &MapBlock{
		Mappings: mappings,
		NodeData: nodeData,
	}, nil
}

func DecodeMapBlock(data []byte) (*MapBlock, error) {
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
