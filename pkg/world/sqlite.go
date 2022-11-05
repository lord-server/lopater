package world

import (
	"database/sql"

	"github.com/lord-server/lopater/pkg/spatial"
)

type SQLiteStorage struct {
	db       *sql.DB
	getBlock *sql.Stmt
	setBlock *sql.Stmt
}

func openSQLite(path string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	getBlock, err := db.Prepare("SELECT data FROM blocks WHERE pos = ?")
	if err != nil {
		return nil, err
	}
	setBlock, err := db.Prepare("INSERT INTO blocks(pos, data) VALUES(?, ?) ON CONFLICT(pos) DO UPDATE SET data = excluded.data")
	if err != nil {
		return nil, err
	}

	return &SQLiteStorage{
		db:       db,
		getBlock: getBlock,
		setBlock: setBlock,
	}, nil
}

func (s *SQLiteStorage) GetBlockData(pos spatial.BlockPosition) ([]byte, error) {
	var data []byte
	err := s.getBlock.QueryRow(pos.Encode()).Scan(&data)

	if err != nil {
		return data, err
	}

	return data, nil
}

func (s *SQLiteStorage) SetBlockData(pos spatial.BlockPosition, data []byte) error {
	panic("unimplemented")
}

func (s *SQLiteStorage) Close() {
	s.db.Close()
}
