package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"log_shelter/internal/infra/reader"
	"log_shelter/internal/infra/repository"
	"log_shelter/internal/usecase"
	"time"

	"github.com/nats-io/nats.go"
)

func ParseInput[T any](data []byte) (*T, error) {
	var r T
	err := json.Unmarshal(data, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *Server) handlerAppendLog(msg *nats.Msg) {
	nc := s.nats.Conn

	nc.Publish("nats.__internal.append", msg.Data)
}
func (s *Server) handlerGetLog(msg *nats.Msg) {
	input, err := ParseInput[usecase.GetLogRequest](msg.Data)
	if err != nil {
		slog.Error("Error while parsing input", "err", err)
		return
	}

	conn, tx, err := s.pg.GetTranscation()
	if err != nil {
		slog.Error("Error before transaction", "err", err)
		return
	}
	defer conn.Close()

	u := usecase.GetLogUsecase{Tx: tx, LogReader: reader.NewLogReader(s.ctx, tx)}

	data, err := u.Run(*input)

	if err != nil {
		tx.Rollback()
		slog.Error("Error in usecase", "err", err)
		return
	}

	err = msg.Respond(data)

	if err != nil {
		tx.Rollback()
		slog.Error("Error in respond", "err", err)
		return
	}
}

func (s *Server) internalAppendHandler(ctx context.Context, sub *nats.Subscription) {
	for {
		select {
		case <-ctx.Done():
			return
		default:

		}

		data, err := sub.FetchBatch(10, nats.MaxWait(1*time.Second))
		if err != nil {
			continue
		}
		i := uint64(0)
		for msg := range data.Messages() {
			input, err := ParseInput[usecase.AppendLogRequest](msg.Data)
			if err != nil {
				slog.Error("Error while parsing input", "err", err)
				return
			}

			conn, tx, err := s.pg.GetTranscation()
			if err != nil {
				slog.Error("Error before transaction", "err", err)
				return
			}
			defer conn.Close()

			u := usecase.AppendLogUsecase{Tx: tx, LogRepo: repository.NewLogRepository(s.ctx, tx)}

			err = u.Run(*input)

			if err != nil {
				tx.Rollback()
				slog.Error("Error in usecase", "err", err)
				return
			}
			i += 1
		}
		slog.Info("Cycle ended", "iters", i)
	}
}

func (s *Server) setupAPI() {
	nc := s.nats.Conn
	js, err := nc.JetStream()

	if err != nil {
		panic(err)
	}

	_, err = nc.Subscribe(
		"nats.hi",
		s.handlerAppendLog,
	)

	if err != nil {
		slog.Default().Error("Cannot create subscriber", "err", err)
	}

	_, err = nc.Subscribe(
		"nats.bye",
		s.handlerGetLog,
	)

	if err != nil {
		slog.Default().Error("Cannot create subscriber", "err", err)
	}

	internal_append_sub, err := js.PullSubscribe("nats.__internal.append", "my_cons")
	if err != nil {
		slog.Default().Error("Cannot create subscriber", "err", err)
	}
	go s.internalAppendHandler(s.ctx, internal_append_sub)
}
