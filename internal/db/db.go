package db

import (
	"context"
	"fmt"
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

// Initializes the option_chain_snapshots table
func InitOptionChainSnapshotsTable(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
	CREATE TABLE IF NOT EXISTS option_chain_snapshots (
		id SERIAL PRIMARY KEY,
		timestamp TIMESTAMPTZ DEFAULT NOW(),
		symbol VARCHAR(50),
		expiry_date DATE NOT NULL,
		strike_price NUMERIC(10, 2) NOT NULL,
		underlying_value NUMERIC(10,2),
		ce_oi BIGINT DEFAULT 0,
		ce_ch_oi BIGINT DEFAULT 0,
		ce_vol BIGINT DEFAULT 0,
		ce_iv NUMERIC(10,2) DEFAULT 0,
		ce_ltp NUMERIC(10,2) DEFAULT 0,
		pe_oi BIGINT DEFAULT 0,
		pe_ch_oi BIGINT DEFAULT 0,
		pe_vol BIGINT DEFAULT 0,
		pe_iv NUMERIC(10,2) DEFAULT 0,
		pe_ltp NUMERIC(10,2) DEFAULT 0,
		intraday_pcr NUMERIC(10,2),
		pcr NUMERIC(10,2)
	);

	CREATE INDEX IF NOT EXISTS idx_option_chain_symbol_expiry
	ON option_chain_snapshots(symbol, expiry_date);
	`

	_, err := pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to initialize option_chain_snapshots table: %w", err)
	}

	return nil
}

// Creates singleton DB instance
func NewDB(ctx context.Context, connString string) (*DB, error) {
	var err error
	pgOnce.Do(func() {
		pool, e := pgxpool.New(ctx, connString)
		if e != nil {
			log.Fatal("Unable to connect to database:", e)
			err = e
			return
		}

		// Initialize table automatically
		if e := InitOptionChainSnapshotsTable(ctx, pool); e != nil {
			log.Fatal("Failed to initialize table:", e)
			err = e
			return
		}

		pgInstance = &DB{db: pool}
	})
	return pgInstance, err
}

// Writes option chain records to DB
func (db *DB) WriteToDB(ctx context.Context, records []models.ResponsePayload) error {
	for _, p := range records {
		_, err := db.db.Exec(ctx, `
			INSERT INTO option_chain_snapshots (
				timestamp, expiry_date, strike_price, underlying_value,
				ce_oi, ce_ch_oi, ce_vol, ce_iv, ce_ltp,
				pe_oi, pe_ch_oi, pe_vol, pe_iv, pe_ltp,
				intraday_pcr, pcr
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
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
