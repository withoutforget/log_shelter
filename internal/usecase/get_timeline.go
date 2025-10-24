package usecase

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"log_shelter/internal/infra/reader"
	"time"
)

type GetTimelineRequest struct {
	ID     uint64         `json:"id"`
	Before *time.Duration `json:"before,omitempty"`
	After  *time.Duration `json:"after,omitempty"`
}

type GetTimelineUsecase struct {
	Tx        *sql.Tx
	LogReader *reader.LogReader
}

func (u *GetTimelineUsecase) Run(data GetTimelineRequest) ([]byte, error) {
	result, err := u.LogReader.GetTimeLineFor(
		data.ID,
		data.Before,
		data.After,
	)
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
