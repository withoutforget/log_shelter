package server

import (
	"log/slog"

	"github.com/nats-io/nats.go"
)

func (s *Server) handlerHi(msg *nats.Msg) {
	logger := slog.Default()

	logger.Info("Got message", "msg", msg)
}

func (s *Server) setupAPI() {
	nc := s.nats.Conn

	_, err := nc.Subscribe(
		"nats.hi",
		s.handlerHi,
	)

	if err != nil {
		slog.Default().Error("Cannot create subscriber", "err", err)
	}
}
