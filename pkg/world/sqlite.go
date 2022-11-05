package world

import (
	"database/sql"

	"github.com/lord-server/lopater/pkg/spatial"
)

type SQLiteStorage struct {
	db          *sql.DB
	getMapBlock *sql.Stmt
	setMapBlock *sql.Stmt
}

func openSQLite(path string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	getMapBlock, err := db.Prepare("SELECT data FROM blocks WHERE pos = ?")
	if err != nil {
		return nil, err
	}
	setMapBlock, err := db.Prepare("INSERT INTO blocks(pos, data) VALUES(?, ?) ON CONFLICT(pos) DO UPDATE SET data = excluded.data")
	if err != nil {
		return nil, err
	}

	return &SQLiteStorage{
		db:          db,
		getMapBlock: getMapBlock,
		setMapBlock: setMapBlock,
	}, nil
}

func (s *SQLiteStorage) GetMapBlockData(pos spatial.MapBlockPosition) ([]byte, error) {
	var data []byte
	err := s.getMapBlock.QueryRow(pos.Encode()).Scan(&data)

	if err != nil {
		return data, err
	}

	return data, nil
}

func (s *SQLiteStorage) SetMapBlockData(pos spatial.MapBlockPosition, data []byte) error {
	panic("unimplemented")
}

func (s *SQLiteStorage) Close() {
	s.db.Close()
}
