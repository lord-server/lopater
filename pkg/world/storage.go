package world

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type Storage interface {
	// EnumerateBlocks()
	GetBlockData(pos Position) ([]byte, error)
	SetBlockData(pos Position, data []byte) error
}

type SQLiteStorage struct {
	db       *sql.DB
	getBlock *sql.Stmt
	setBlock *sql.Stmt
}

func openSQLite(path string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", "./foo.db")
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

func (s *SQLiteStorage) GetBlockData(pos Position) ([]byte, error) {
	return []byte{}, nil
}

func (s *SQLiteStorage) SetBlockData(pos Position, data []byte) error {
	return nil
}
