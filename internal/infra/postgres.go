package infra

import (
	"context"
	"log_shelter/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresInfra struct {
	Pool *pgxpool.Pool
}

func NewPostgresInfra(ctx context.Context, cfg *config.PostgresConfig) (*PostgresInfra, error) {
	var ret PostgresInfra

	pg_cfg, err := pgxpool.ParseConfig(cfg.Dsn())
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, pg_cfg)

	if err != nil {
		return nil, err
	}

	ret.Pool = pool

	return &ret, nil
}
