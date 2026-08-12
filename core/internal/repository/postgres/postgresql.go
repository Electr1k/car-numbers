package postgres

import (
	"context"
	"core/config"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, config config.DatabaseConfig) (*Postgres, error) {
	pool, err := pgxpool.Connect(ctx, config.URL)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return &Postgres{pool: pool}, nil
}

func (p *Postgres) Close() {
	if p.pool != nil {
		p.pool.Close()
	}
}
