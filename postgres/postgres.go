package postgres

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	_defaultConnAttempts = 3
	_defaultConnTimeout  = time.Second
)

type DBConnString string

type postgres struct {
	connAttempts int
	connTimeout  time.Duration

	db *pgxpool.Pool
}

var _ DBEngine = (*postgres)(nil)

func NewPostgresDB(url DBConnString) (DBEngine, error) {
	pg := &postgres{
		connAttempts: _defaultConnAttempts,
		connTimeout:  _defaultConnTimeout,
	}

	var err error
	for pg.connAttempts > 0 {
		pg.db, err = pgxpool.New(context.Background(), string(url))
		if err == nil {
			slog.Info("📰 connected to Postgres 🎉")
			return pg, nil
		}

		pg.connAttempts--
		log.Printf("❌ connect failed: %v – retrying (%d left)", err, pg.connAttempts)
		time.Sleep(pg.connTimeout)
	}

	return nil, fmt.Errorf("connect attempts exceeded (%d): %w",
		pg.connAttempts, err)
}

func (p *postgres) Configure(opts ...Option) DBEngine {
	for _, opt := range opts {
		opt(p)
	}

	return p
}

func (p *postgres) GetDB() *pgxpool.Pool {
	return p.db
}

func (p *postgres) Close() {
	if p.db != nil {
		p.db.Close()
	}
}
