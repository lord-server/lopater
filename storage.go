package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

type BlockDataHandler func(x, y, z int, data []byte) error

type WorldStorage interface {
	GetBlockData(ctx context.Context, x, y, z int) ([]byte, error)
	GetBlocksData(ctx context.Context, region Region, cb BlockDataHandler) error
	SetBlockData(ctx context.Context, x, y, z int, data []byte) error
}

type PgStorage struct {
	pool *pgxpool.Pool
}

func newPgStorage(ctx context.Context, connString string) (*PgStorage, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, err
	}

	return &PgStorage{
		pool: pool,
	}, nil
}

func (s *PgStorage) Close() {
	s.pool.Close()
}

func (s *PgStorage) GetBlockData(ctx context.Context, x, y, z int) ([]byte, error) {
	const sql = `select data
	             from blocks
	             where posx=$1
	               and posy=$2
	               and posz=$3`

	row := s.pool.QueryRow(ctx, sql, x, y, z)

	var blockData []byte

	err := row.Scan(&blockData)
	if err != nil {
		return nil, err
	}

	return blockData, nil
}

func (s *PgStorage) GetBlocksData(ctx context.Context, region Region, handler BlockDataHandler) error {
	const sql = `select posx, posy, posz, data
	             from blocks
	             where posx between $1 and $2
	               and posy between $3 and $4
	               and posz between $5 and $6`

	rows, err := s.pool.Query(ctx, sql,
		region.MinX,
		region.MaxX,
		region.MinY,
		region.MaxY,
		region.MinZ,
		region.MaxZ)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var (
			x, y, z int
			data    []byte
		)

		err = rows.Scan(&x, &y, &z, &data)
		if err != nil {
			return err
		}

		err = handler(x, y, z, data)
		if err != nil {
			log.Printf("block %v,%v,%v: %v, %v", x, y, z, err, data)
			continue
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}

func (s *PgStorage) SetBlockData(ctx context.Context, x, y, z int, data []byte) error {
	panic("unimplemented")
}
