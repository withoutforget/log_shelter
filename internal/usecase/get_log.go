package usecase

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"time"

	"log_shelter/internal/infra/reader"
)

type GetLogRequest struct {
	Page       uint64     `json:"page"`
	PageSize   *uint64    `json:"page_size,omitempty"`
	Sources    []string   `json:"sources,omitempty"`
	Levels     []string   `json:"levels,omitempty"`
	Before     *time.Time `json:"before,omitempty"`
	After      *time.Time `json:"after,omitempty"`
	RequestID  *string    `json:"request_id,omitempty"`
	LoggerName *string    `json:"logger_name,omitempty"`
	Order      string     `json:"order"`
}

type GetLogUsecase struct {
	Tx        *sql.Tx
	LogReader *reader.LogReader
}

func (u *GetLogUsecase) Run(data GetLogRequest) ([]byte, error) {
	result, err := u.LogReader.ReadLogs(data.Page,
		data.PageSize,
		data.Sources,
		data.Levels,
		data.Before,
		data.After,
		data.RequestID,
		data.LoggerName,
		reader.OrderT(data.Order))
	if err != nil {
		u.Tx.Rollback()
		slog.Error("oops... read", "Err", err)
	}
	bytes, err := json.Marshal(result)
	if err != nil {
		u.Tx.Rollback()
		slog.Error("oops... to json", "Err", err)
	}
	u.Tx.Commit()
	return bytes, err
}
