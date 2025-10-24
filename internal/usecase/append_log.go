package usecase

import (
	"database/sql"
	"log/slog"
	"time"

	"log_shelter/internal/infra/repository"
)

type AppendLogRequest struct {
	RawLog     string    `json:"raw_log"`
	LogLevel   string    `json:"log_level"`
	Source     string    `json:"source"`
	CreatedAt  time.Time `json:"created_at"`
	RequestID  *string   `json:"request_id,omitempty"`
	LoggerName *string   `json:"logger_name,omitempty"`
}

type AppendLogUsecase struct {
	Tx      *sql.Tx
	LogRepo *repository.LogRepository
}

func (u *AppendLogUsecase) Run(data AppendLogRequest) error {
	err := u.LogRepo.AppendLog(
		data.RawLog,
		data.LogLevel,
		data.Source,
		data.CreatedAt,
		data.RequestID,
		data.LoggerName,
	)
	if err != nil {
		u.Tx.Rollback()
		slog.Error("oops... append", "Err", err)
	}
	u.Tx.Commit()
	return err
}
