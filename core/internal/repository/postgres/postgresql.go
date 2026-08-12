package postgres

import (
	"context"
	"core/config"
	"core/pkg/utils"

	"github.com/jackc/pgx/v4/pgxpool"
)

const databaseName = "postgres"

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

	if err := utils.RunMigrations(ctx, pool, databaseName); err != nil {
		return nil, err
	}

	return &Postgres{pool: pool}, nil
}

func (p *Postgres) Close() {
	if p.pool != nil {
		p.pool.Close()
	}
}
