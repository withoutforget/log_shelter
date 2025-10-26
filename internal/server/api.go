package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
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

func (s *Server) handlerDebezium(msg *nats.Msg) {
	err := s.es.Handle(msg.Data)

	if err != nil {
		slog.Error("Error", "error", err)
	}
}

func (s *Server) handlerSearch(resp http.ResponseWriter, req *http.Request) {
	v, e := req.URL.Query()["q"]
	if !e || len(v) == 0 {
		http.Error(resp, "Invalid \"q\" param", 400)
		return
	}
	ret := make([]map[string]any, 0)
	for _, v := range s.es.Search(v[0]) {
		tmp := map[string]any{}
		err := json.Unmarshal([]byte(v), &tmp)
		if err != nil {
			http.Error(resp, "Err", 500)
			return

		}
		ret = append(ret, tmp)
	}

	resp.Header().Set("Content-Type", "application/json")
	json.NewEncoder(resp).Encode(ret)
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

	internal_append_sub, err := js.PullSubscribe("nats.__internal.append", "append_stream")
	if err != nil {
		slog.Default().Error("Cannot create subscriber", "err", err)
	}
	go s.internalAppendHandler(s.ctx, internal_append_sub)
	go s.logRetention(s.ctx)

	_, err = nc.Subscribe(
		"postgres.*.*",
		s.handlerDebezium,
	)
	if err != nil {
		slog.Default().Error("Cannot create subscriber", "err", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/search", s.handlerSearch)
	srv := &http.Server{Addr: "0.0.0.0:80", Handler: mux}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			slog.Error("Http server closed: %v", "error", err)
		}
	}()

	go func() {
		<-s.ctx.Done()
		srv.Shutdown(s.ctx)
	}()
}
