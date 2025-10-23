package server

import (
	"context"
	"log_shelter/internal/config"
	"log_shelter/internal/infra"
)

type Server struct {
	ctx  context.Context
	cfg  *config.Config
	nats *infra.NatsInfra
	pg   *infra.PostgresInfra
}

func NewServer(ctx context.Context, cfg *config.Config) *Server {
	var srv Server

	srv.cfg = cfg
	nats, err := infra.NewNatsInfra(&cfg.Nats)

	if err != nil {
		panic(err)
	}

	pg, err := infra.NewPostgresInfra(ctx, &cfg.Postgres)

	if err != nil {
		panic(err)
	}

	srv.nats = nats
	srv.ctx = ctx
	srv.pg = pg

	return &srv
}

func (s *Server) Run() {
	s.setupAPI()

	<-s.ctx.Done()
}
