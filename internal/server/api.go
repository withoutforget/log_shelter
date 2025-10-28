package server

import (
	"context"
	"log/slog"
	"time"

	"log_shelter/internal/infra/repository"
)

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
	s.setupNatsAPI()
	s.setupHTTPAPI()

	go s.logRetention(s.ctx)
}
