package postgres

import (
	"context"
	"core/pkg/utils"
	"database/sql"
	"sync"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Postgres struct {
	mu   sync.Mutex
	pool *pgxpool.Pool
}

func New(ctx context.Context, connectionString string) (*Postgres, error) {
	pool, err := pgxpool.Connect(ctx, connectionString)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	if err := utils.RunMigrations(ctx, pool, "postgres"); err != nil {
		return nil, err
	}

	return &Postgres{pool: pool}, nil
}

func (p *Postgres) Close() {
	if p.pool != nil {
		p.pool.Close()
	}
}

func nullFloatToFloat32Ptr(n sql.NullFloat64) *float32 {
	if n.Valid {
		val := float32(n.Float64)
		return &val
	}
	return nil
}
