package infra

import (
	"log/slog"
	"log_shelter/internal/config"
	"sync/atomic"

	"github.com/nats-io/nats.go"
)

type NatsStatus int

const (
	NatsStatusOK NatsStatus = iota
	NatsStatusDisconnected
	NatsStatusClosed
)

type NatsInfra struct {
	Conn   *nats.Conn
	Status atomic.Int32
}

func NewNatsInfra(cfg *config.NatsConfig) (*NatsInfra, error) {
	var ret NatsInfra

	disconnect_handler := func(nc *nats.Conn, err error) {
		logger := slog.Default()

		logger.Error("Nats has been disconnected",
			slog.String("url", nc.Opts.Url),
			slog.String("username", nc.Opts.User),
			slog.String("error", err.Error()),
		)

		ret.Status.Store(int32(NatsStatusDisconnected))
	}

	close_handler := func(nc *nats.Conn) {
		logger := slog.Default()

		logger.Error("Nats has been closed",
			slog.String("url", nc.Opts.Url),
			slog.String("username", nc.Opts.User),
		)

		ret.Status.Store(int32(NatsStatusClosed))
	}

	conn, err := nats.Connect(
		cfg.Url,
		nats.UserInfo(cfg.Username, cfg.Password),
		nats.DisconnectErrHandler(disconnect_handler),
		nats.ClosedHandler(close_handler),
	)
	if err != nil {
		return nil, err
	}

	ret.Conn = conn

	return &ret, nil
}
