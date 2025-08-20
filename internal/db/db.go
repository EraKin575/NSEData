package db

import (
	"context"
	"log"
	"server/models"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	db *pgxpool.Pool
}

var (
	pgInstance *DB
	pgOnce     sync.Once
)

func NewDB(ctx context.Context, connString string) (*DB, error) {
	pgOnce.Do(func() {
		db, err := pgxpool.New(ctx, connString)
		if err != nil {
			log.Fatal("Unable to connect to database:", err)
		}
		pgInstance = &DB{
			db: db,
		}
	})
	return pgInstance, nil
}

func (db *DB) WriteToDB(ctx context.Context, records []models.ResponsePayload) error {

	for _, p := range records {
		_, err := db.db.Exec(ctx, `
			INSERT INTO option_chain_snapshots (
				timestamp, expiry_date, strike_price, underlying_value,
				ce_oi, ce_ch_oi, ce_vol, ce_iv, ce_ltp,
				pe_oi, pe_ch_oi, pe_vol, pe_iv, pe_ltp,
				intraday_pcr, pcr, signal
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
		`,
			p.Timestamp, p.ExpiryDate, p.StrikePrice, p.UnderlyingValue,
			p.CEOpenInterest, p.CEChangeInOpenInterest, p.CETotalTradedVolume, p.CEImpliedVolatility, p.CELastPrice,
			p.PEOpenInterest, p.PEChangeInOpenInterest, p.PETotalTradedVolume, p.PEImpliedVolatility, p.PELastPrice,
			p.IntraDayPCR, p.PCR,
		)
		if err != nil {
			log.Printf("❌ Failed to insert row: %v", err)
		}
	}

	return nil
}
