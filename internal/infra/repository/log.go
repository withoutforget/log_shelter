package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/Masterminds/squirrel"
)

type LogRepository struct {
	ctx context.Context
	tx  *sql.Tx
}

func NewLogRepository(
	ctx context.Context,
	tx *sql.Tx) *LogRepository {
	return &LogRepository{tx: tx, ctx: ctx}
}

func (r *LogRepository) AppendLog(
	raw_log string,
	log_level string,
	source string,
	created_at time.Time,
	request_id *string,
	logger_name *string,
) error {
	q, args, err := squirrel.Insert("logs").Columns(
		"raw_log",
		"log_level",
		"source",
		"created_at",
		"request_id",
		"logger_name",
		"is_deleted",
	).Values(
		raw_log,
		log_level,
		source,
		created_at,
		request_id,
		logger_name,
		false,
	).PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return err
	}
	_, err = r.tx.Query(q, args...)

	return err
}

func (r *LogRepository) RetentOlder(delta time.Duration) error {
	q := `
		UPDATE logs
		SET is_deleted = true
		WHERE created_at < $1
		AND is_deleted = false
		RETURNING id, created_at, is_deleted;
	`
	deletion_time := time.Now().UTC().Add(-delta)
	_, err := r.tx.Query(q, deletion_time)

	return err
}
