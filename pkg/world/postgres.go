package world

import (
	"context"
	"errors"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/lord-server/lopater/pkg/spatial"
)

type PostgresMapStorage struct {
	conn *pgxpool.Pool
}

func openPostgres(params string) (*PostgresMapStorage, error) {
	conn, err := pgxpool.Connect(context.Background(), params)
	if err != nil {
		return nil, err
	}

	return &PostgresMapStorage{
		conn,
	}, nil
}

func (s *PostgresMapStorage) Close() {
	s.conn.Close()
}

func (s *PostgresMapStorage) GetMapBlockData(pos spatial.MapBlockPosition) ([]byte, error) {
	var data []byte
	const query = "SELECT data FROM blocks WHERE posx=$1 and posy=$2 and posz=$3"
	err := s.conn.QueryRow(context.Background(), query, pos.X, pos.Y, pos.Z).Scan(&data)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return data, nil
}

func (s *PostgresMapStorage) SetMapBlockData(pos spatial.MapBlockPosition, data []byte) error {
	const query = `
	INSERT INTO blocks(posx, posy, posz, data)
		VALUES($1, $2, $3, $4)
		ON CONFLICT(posx, posy, posz) DO
			UPDATE SET data = EXCLUDED.data
	`
	_, err := s.conn.Exec(context.Background(), query, pos.X, pos.Y, pos.Z, data)

	return err
}
