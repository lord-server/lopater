package mapblock

import (
	"fmt"
	"io"

	"github.com/klauspost/compress/zstd"
)

func readStaticObject(reader *binaryReader) (StaticObject, error) {
	var object StaticObject
	var err error

	object.Type, err = reader.ReadUint8()
	if err != nil {
		return object, err
	}

	object.X, err = reader.ReadInt32()
	if err != nil {
		return object, err
	}

	object.Y, err = reader.ReadInt32()
	if err != nil {
		return object, err
	}

	object.Z, err = reader.ReadInt32()
	if err != nil {
		return object, err
	}

	dataSize, err := reader.ReadUint16()
	if err != nil {
		return object, err
	}

	object.Data = make([]byte, dataSize)
	_, err = io.ReadFull(reader, object.Data)
	if err != nil {
		return object, err
	}

	return object, nil
}

func readNodeTimer(reader *binaryReader) (NodeTimer, error) {
	var timer NodeTimer
	var err error

	timer.Position, err = reader.ReadUint16()
	if err != nil {
		return timer, err
	}

	timer.Timeout, err = reader.ReadInt32()
	if err != nil {
		return timer, err
	}

	timer.Elapsed, err = reader.ReadInt32()
	if err != nil {
		return timer, err
	}

	return timer, nil
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

func readStaticObjects(reader *binaryReader) ([]StaticObject, error) {
	var staticObjects []StaticObject

	staticObjectCount, err := reader.ReadUint16()
	if err != nil {
		return nil, fmt.Errorf("unable to decode static object count: %w", err)
	}

	for i := 0; i < int(staticObjectCount); i++ {
		object, err := readStaticObject(reader)
		if err != nil {
			return nil, err
		}
		staticObjects = append(staticObjects, object)
	}

	return staticObjects, nil
}

func readNodeTimers(reader *binaryReader) ([]NodeTimer, error) {
	var nodeTimers []NodeTimer
	nodeTimerCount, err := reader.ReadUint16()
	if err != nil {
		return nil, fmt.Errorf("unable to decode node timer count: %w", err)
	}

	for i := 0; i < int(nodeTimerCount); i++ {
		nodeTimer, err := readNodeTimer(reader)
		if err != nil {
			return nil, fmt.Errorf("unable to decode node timer: %w", err)
		}

		nodeTimers = append(nodeTimers, nodeTimer)
	}

	return nodeTimers, nil
}

// decodeLegacyMapBlock decodes zlib-compressed MapBlocks (Minetest versions before 5.5)
func decodeLegacyMapBlock(reader *binaryReader, version uint8) (*MapBlock, error) {
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
	// - uint8 content_width = 2
	// - uint8 params_width = 2
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

	mapBlock.StaticObjects, err = readStaticObjects(reader)
	if err != nil {
		return nil, fmt.Errorf("unable to decode static objects: %w", err)
	}

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

	// Skip constant value:
	// - uint8 nodeTimerSize
	_, err = reader.Seek(1, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	mapBlock.NodeTimers, err = readNodeTimers(reader)
	if err != nil {
		return nil, err
	}

	return &mapBlock, nil
}

// decodeMapBlock decodes MapBlocks zstd-compressed MapBlocks (Minetest 5.5 onwards)
func decodeMapBlock(reader *binaryReader) (*MapBlock, error) {
	zstdReader, err := zstd.NewReader(reader)
	if err != nil {
		return nil, fmt.Errorf("unable to create zstd decoder: %w", err)
	}
	defer zstdReader.Close()

	data, err := io.ReadAll(zstdReader)
	if err != nil {
		return nil, fmt.Errorf("unable to read data from zstd stream: %w", err)
	}

	var mapBlock MapBlock
	reader = newBinaryReader(data)

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
		return nil, fmt.Errorf("unable to decode lighting flags: %w", err)
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

	// Skip constant values:
	// - uint8 content_width = 2
	// - uint8 params_width = 2
	_, err = reader.Seek(1+1, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	mapBlock.NodeData = make([]byte, MapBlockVolume*NodeSizeInBytes)
	_, err = io.ReadFull(reader, mapBlock.NodeData)
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

	if version < MinSupportedVersion || version > MaxSupportedVersion {
		return nil, fmt.Errorf("unsupported MapBlock version: %v", version)
	}

	if version < 29 {
		mapblock, err := decodeLegacyMapBlock(reader, version)
		if err != nil {
			return nil, err
		}
		return mapblock, nil
	}

	return decodeMapBlock(reader)
}
