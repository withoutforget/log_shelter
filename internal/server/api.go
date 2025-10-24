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
		timer := time.NewTimer(time.Duration(s.cfg.Logs.CycleTime))
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}

		data, err := sub.FetchBatch(100, nats.MaxWait(100*time.Millisecond))
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

func (s *Server) logRetention(ctx context.Context) {
	cfg := s.cfg.Logs
	pg := s.pg
	for {
		time.Sleep(time.Duration(cfg.CycleTime))
		select {
		case <-ctx.Done():
			return
		default:

		}

		conn, tx, err := pg.GetTranscation()
		if err != nil {
			slog.Error("Error before transcation in retention", "err", err)
		}

		defer conn.Close()

		repo := repository.NewLogRepository(ctx, tx)

		switch cfg.RetencionPolicy {
		case "after_time":
			err := repo.RetentOlder(time.Duration(cfg.DeleteAfter))
			if err != nil {
				slog.Error("Error in retention", "err", err)
			}
			err = tx.Commit()
			if err != nil {
				slog.Error("Error in retention commit", "err", err)
			}
		default:
			tx.Rollback()
			return
		}
		slog.Info("Cycle ended")
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
		"nats.timeline",
		s.handlerGetTimeline,
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
	go s.logRetention(s.ctx)
}
