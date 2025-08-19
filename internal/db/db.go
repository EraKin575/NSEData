package db

import (
	"context"
	"log"
	"server/models"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgres struct {
	db *pgxpool.Pool
}

var (
	pgInstance *postgres
	pgOnce     sync.Once
)

func NewDB(ctx context.Context, connString string) (*postgres, error) {
	pgOnce.Do(func() {
		db, err := pgxpool.New(ctx, connString)
		if err != nil {
			log.Fatal("Unable to connect to database:", err)
		}
		pgInstance = &postgres{
			db: db,
		}
	})
	return pgInstance, nil
}


