package world

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type BackendType int

const (
	BackendSQLite BackendType = iota
	BackendPostgreSQL
)

type WorldMetadata struct {
	BackendType BackendType
}

func parseBackend(backend string) (BackendType, error) {
	switch backend {
	case "sqlite3":
		return BackendSQLite, nil
	case "postgresql":
		return BackendPostgreSQL, nil
	default:
		return BackendSQLite, fmt.Errorf("unknown storage backend: %v", backend)
	}
}

func (m *WorldMetadata) parseLine(line string) error {
	parts := strings.SplitN(line, "=", 2)

	if len(parts) != 2 {
		return fmt.Errorf("invalid line `%v`", line)
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	var err error

	// FIXME: return error if required fields weren't found
	switch key {
	case "backend":
		m.BackendType, err = parseBackend(value)
		if err != nil {
			return err
		}
	}

	return nil
}

func ReadMetadata(worldMtPath string) (WorldMetadata, error) {
	var metadata WorldMetadata

	file, err := os.Open(worldMtPath)
	if err != nil {
		return metadata, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		metadata.parseLine(line)
	}

	return metadata, nil
}
