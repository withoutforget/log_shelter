package infra

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"

	"log_shelter/internal/config"
)

type PostgresInfra struct {
	ctx context.Context
	cfg *config.PostgresConfig
	db  *sql.DB
}

func NewPostgresInfra(ctx context.Context, cfg *config.PostgresConfig) (*PostgresInfra, error) {
	var ret PostgresInfra
	ret.cfg = cfg
	ret.ctx = ctx
	return &ret, nil
}

func (p *PostgresInfra) GetConnection() (*sql.Conn, error) {
	if p.db == nil {
		pg, err := sql.Open("postgres", p.cfg.Dsn())
		if err == nil {
			return nil, err
		}
		p.db = pg
	}
	return p.db.Conn(p.ctx)
}

func (p *PostgresInfra) GetTranscation() (*sql.Conn, *sql.Tx, error) {
	conn, err := p.GetConnection()
	if err != nil {
		return nil, nil, err
	}
	tx, err := conn.BeginTx(p.ctx, nil)
	if err != nil {
		conn.Close()
		return nil, nil, err
	}
	return conn, tx, nil
}
