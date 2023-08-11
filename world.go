package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrInvalidContentWidth     = errors.New("block has invalid content width")
	ErrInvalidParamWidth       = errors.New("block has invalid param width")
	ErrInvalidStaticObjVersion = errors.New("block has invalid static object version")
	ErrInvalidMappingVersion   = errors.New("block has invalid mapping version")
	ErrInvalidTimerDataLength  = errors.New("block has invalid timer data size")
)

type World struct {
	storage  WorldStorage
	metadata WorldMetadata
}

type WorldMetadata struct {
	backend      string
	pgConnString string
}

func (m *WorldMetadata) parseLine(line string) {
	line = strings.TrimSpace(line)

	parts := strings.SplitN(line, "=", 2)

	if len(parts) < 2 {
		return
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	switch key {
	case "backend":
		m.backend = value
	case "pgsql_connection":
		m.pgConnString = value
	}
}

func parseWorldMetadata(path string) (meta WorldMetadata, err error) {
	file, err := os.Open(path)
	if err != nil {
		return meta, err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		meta.parseLine(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return meta, err
	}

	return meta, err
}

func OpenWorld(ctx context.Context, path string) (*World, error) {
	worldMetadataPath := filepath.Join(path, "world.mt")

	metadata, err := parseWorldMetadata(worldMetadataPath)
	if err != nil {
		return nil, err
	}

	var storage WorldStorage

	switch metadata.backend {
	case "postgresql":
		storage, err = newPgStorage(ctx, metadata.pgConnString)
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unknown backend: %v", metadata.backend)
	}

	return &World{
		storage:  storage,
		metadata: metadata,
	}, nil
}

func (w *World) GetBlock(ctx context.Context, x, y, z int) (*Block, error) {
	data, err := w.storage.GetBlockData(ctx, x, y, z)
	if err != nil {
		return nil, err
	}

	return DecodeBlock(data)
}

type Block struct {
	Flags            uint8
	LightingComplete uint16
	NodeData         []byte
	Mapping          map[uint16]string
	StaticObjects    []StaticObject
	Timestamp        uint32
	NodeTimers       []NodeTimer
}

func DecodeBlock(rawData []byte) (*Block, error) {
	r := newReader(rawData)

	version, err := r.readUint8()
	if err != nil {
		return nil, err
	}

	switch {
	case version >= 25 && version <= 28:
		return decodeZlibBlock(version, r)
	case version >= 29:
		return decodeZstdBlock(version, r)
	}

	return nil, fmt.Errorf("unsupported block version: %v", version)
}

func decodeZlibBlock(version uint8, r *reader) (block *Block, err error) {
	block = new(Block)

	block.Flags, err = r.readUint8()
	if err != nil {
		return nil, err
	}

	if version >= 27 {
		block.LightingComplete, err = r.readUint16()
		if err != nil {
			return nil, err
		}
	}

	contentWidth, err := r.readUint8()
	if err != nil {
		return nil, err
	}

	if contentWidth != 2 {
		return nil, ErrInvalidContentWidth
	}

	paramWidth, err := r.readUint8()
	if err != nil {
		return nil, err
	}

	if paramWidth != 2 {
		return nil, ErrInvalidContentWidth
	}

	block.NodeData, err = r.readZlib()
	if err != nil {
		return nil, err
	}

	_, err = r.readZlib()
	if err != nil {
		return nil, err
	}

	staticObjectVersion, err := r.readUint8()
	if err != nil {
		return nil, err
	}

	if staticObjectVersion != 0 {
		return nil, ErrInvalidStaticObjVersion
	}

	staticObjectCount, err := r.readUint16()
	if err != nil {
		return nil, err
	}

	for i := 0; i < int(staticObjectCount); i++ {
		staticObject, err := readStaticObject(r)
		if err != nil {
			return nil, err
		}

		block.StaticObjects = append(block.StaticObjects, staticObject)
	}

	block.Timestamp, err = r.readUint32()
	if err != nil {
		return nil, err
	}

	nameIDMappingVersion, err := r.readUint8()
	if err != nil {
		return nil, err
	}

	if nameIDMappingVersion != 0 {
		return nil, ErrInvalidMappingVersion
	}

	block.Mapping = make(map[uint16]string)

	mappingCount, err := r.readUint16()
	if err != nil {
		return nil, err
	}

	for i := 0; i < int(mappingCount); i++ {
		nodeID, err := r.readUint16()
		if err != nil {
			return nil, err
		}

		name, err := r.readString()
		if err != nil {
			return nil, err
		}

		block.Mapping[nodeID] = name
	}

	timerDataLength, err := r.readUint8()
	if err != nil {
		return nil, err
	}

	if timerDataLength != 10 {
		return nil, ErrInvalidTimerDataLength
	}

	timerCount, err := r.readUint16()
	if err != nil {
		return nil, err
	}

	for i := 0; i < int(timerCount); i++ {
		nodeTimer, err := readNodeTimer(r)
		if err != nil {
			return nil, err
		}

		block.NodeTimers = append(block.NodeTimers, nodeTimer)
	}

	return block, nil
}

func decodeZstdBlock(version uint8, r *reader) (*Block, error) {
	panic("unimplemented")
}

type StaticObject struct {
	Type    uint8
	X, Y, Z int32
	Data    []byte
}

func readStaticObject(r *reader) (StaticObject, error) {
	var (
		object StaticObject
		err    error
	)

	object.Type, err = r.readUint8()
	if err != nil {
		return object, err
	}

	x, err := r.readUint32()
	if err != nil {
		return object, err
	}
	object.X = int32(x)

	y, err := r.readUint32()
	if err != nil {
		return object, err
	}
	object.Y = int32(y)

	z, err := r.readUint32()
	if err != nil {
		return object, err
	}
	object.Z = int32(z)

	object.Data, err = r.readByteSlice()
	if err != nil {
		return object, err
	}

	return object, nil
}

type NodeTimer struct {
	Position uint16
	Timeout  int32
	Elapsed  int32
}

func readNodeTimer(r *reader) (NodeTimer, error) {
	position, err := r.readUint16()
	if err != nil {
		return NodeTimer{}, err
	}

	timeout, err := r.readUint32()
	if err != nil {
		return NodeTimer{}, err
	}

	elapsed, err := r.readUint32()
	if err != nil {
		return NodeTimer{}, err
	}

	return NodeTimer{
		Position: position,
		Timeout:  int32(timeout),
		Elapsed:  int32(elapsed),
	}, nil
}
