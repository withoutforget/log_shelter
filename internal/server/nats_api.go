package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"

	"log_shelter/internal/infra/notifications"
	"log_shelter/internal/infra/reader"
	"log_shelter/internal/infra/repository"
	"log_shelter/internal/usecase"
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
	f, err := s.factory.GetUsecaseFactory(s.ctx)
	if err != nil {
		slog.Error("Cannot get factory", "err", err)
		return
	}
	data, err := f.GetGetLogUsecase().Run(*input)
	if err != nil {
		return
	}
	err = msg.Respond(data)
	if err != nil {
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

		data, err := sub.FetchBatch(100, 100*nats.MaxWait(time.Millisecond))
		if err != nil {
			continue
		}
		i := uint64(0)
		for msg := range data.Messages() {
			err = msg.Ack()
			if err != nil {
				slog.Error("Cannot ACK message in batch", "err", err)
			}
			input, err := ParseInput[usecase.AppendLogRequest](msg.Data)
			if err != nil {
				slog.Error("Error while parsing input", "err", err)
				continue
			}

			conn, tx, err := s.pg.GetTranscation()
			if err != nil {
				slog.Error("Error before transaction", "err", err)
				continue
			}

			u := usecase.AppendLogUsecase{Tx: tx, LogRepo: repository.NewLogRepository(s.ctx, tx)}

			err = u.Run(*input)
			if err != nil {
				e := tx.Rollback()
				conn.Close()
				if e != nil {
					slog.Error("Cannot rollback", "err", e)
				}
				slog.Error("Error in usecase", "err", err)
				continue
			}
			conn.Close()
			if s.tg.ShouldNotify(input.LogLevel) {
				s.tg.Notify(notifications.NotifyLogModel{
					RawLog:     input.RawLog,
					LogLevel:   input.LogLevel,
					Source:     input.Source,
					RequestID:  input.RequestID,
					LoggerName: input.LoggerName,
				})
			}
			i += 1
		}
		if i != 0 {
			slog.Info("Cycle ended", "iters", i)
		}
	}
}

func (s *Server) handlerGetTimeline(msg *nats.Msg) {
	input, err := ParseInput[usecase.GetTimelineRequest](msg.Data)
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

	u := usecase.GetTimelineUsecase{Tx: tx, LogReader: reader.NewLogReader(s.ctx, tx)}

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

func (s *Server) handlerDebezium(msg *nats.Msg) {
	err := s.es.Handle(msg.Data)
	if err != nil {
		slog.Error("Error", "error", err)
	}
}

func (s *Server) setupNatsAPI() {
	nc := s.nats.Conn
	js, err := nc.JetStream()
	if err != nil {
		panic(err)
	}

	_, err = nc.Subscribe(
		"log_shelter.append",
		s.handlerAppendLog,
	)
	if err != nil {
		slog.Default().Error("Cannot create subscriber", "err", err)
	}

	_, err = nc.Subscribe(
		"log_shelter.timeline",
		s.handlerGetTimeline,
	)
	if err != nil {
		slog.Default().Error("Cannot create subscriber", "err", err)
	}

	_, err = nc.Subscribe(
		"log_shelter.get",
		s.handlerGetLog,
	)
	if err != nil {
		slog.Default().Error("Cannot create subscriber", "err", err)
	}

	_, err = nc.Subscribe(
		"log_shelter.__internal.postgres.*.*",
		s.handlerDebezium,
	)
	if err != nil {
		slog.Default().Error("Cannot create subscriber", "err", err)
	}

	internal_append_sub, err := js.PullSubscribe("log_shelter.__internal.append", "append_stream")
	if err != nil {
		slog.Default().Error("Cannot create subscriber", "err", err)
	}
	go s.internalAppendHandler(s.ctx, internal_append_sub)
}
