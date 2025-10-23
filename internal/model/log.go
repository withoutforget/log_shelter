package model

import (
	"encoding/json"
	"time"
)

type LogModel struct {
	ID         uint64    `json:"id"`
	RawLog     string    `json:"raw_log"`
	LogLevel   string    `json:"log_level"`
	Source     string    `json:"source"`
	CreatedAt  time.Time `json:"created_at"`
	RequestID  *string   `json:"request_id"`
	LoggerName string    `json:"logger_name"`
	IsDeleted  bool      `json:"is_deleted"`
}

func (m *LogModel) AsJson() *string {
	res, err := json.Marshal(m)
	if err != nil {
		return nil
	}
	ret := string(res)
	return &ret
}
