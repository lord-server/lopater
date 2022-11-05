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
	Variables   map[string]string
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

	m.Variables[key] = value

	return nil
}

func ReadMetadata(worldMtPath string) (WorldMetadata, error) {
	// FIXME: this parser assumes that world.mt files are well-formed
	// and not malicious

	metadata := WorldMetadata{
		BackendType: BackendSQLite,
		Variables:   make(map[string]string),
	}

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

	if backend, ok := metadata.Variables["backend"]; ok {
		backendType, err := parseBackend(backend)
		if err != nil {
			return metadata, err
		}
		metadata.BackendType = backendType
	} else {
		return metadata, fmt.Errorf("world metadata doesn't specify backend")
	}

	return metadata, nil
}
