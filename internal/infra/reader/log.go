package reader

import (
	"context"
	"database/sql"
	"log_shelter/internal/model"
	"slices"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/lib/pq"
)

type OrderT string

const (
	OrderAsc  OrderT = "asc"
	OrderDesc OrderT = "desc"
)

type LogReader struct {
	ctx context.Context
	tx  *sql.Tx
}

func NewLogReader(
	ctx context.Context,
	tx *sql.Tx) *LogReader {
	return &LogReader{tx: tx, ctx: ctx}
}

func (r *LogReader) ReadLogs(
	page uint64,
	page_size *uint64,
	sources []string,
	levels []string,
	before *time.Time,
	after *time.Time,
	request_id *string,
	logger_name *string,
	order OrderT,
) ([]model.LogModel, error) {
	q := squirrel.Select("id", "raw_log",
		"log_level",
		"source",
		"created_at",
		"request_id",
		"logger_name").From("logs")

	q = q.Where(squirrel.Eq{"is_deleted": false})

	if !slices.Contains(sources, "*") {
		q = q.Where(squirrel.Eq{"source": sources})
	}
	if !slices.Contains(levels, "*") {
		q = q.Where(squirrel.Eq{"level": levels})
	}

	if after != nil {
		q = q.Where(squirrel.GtOrEq{"created_at": *after})
	}
	if before != nil {
		q = q.Where(squirrel.GtOrEq{"created_at": *before})
	}

	if request_id != nil {
		q = q.Where(squirrel.Eq{"request_id": *request_id})
	}
	if logger_name != nil {
		q = q.Where(squirrel.Eq{"logger_name": *logger_name})
	}

	q = q.OrderBy("created_at " + string(order))

	query, args, err := q.PlaceholderFormat(squirrel.Dollar).ToSql()

	if err != nil {
		return nil, err
	}

	rows, err := r.tx.Query(query, args...)

	if err != nil {
		return nil, err
	}

	ret := make([]model.LogModel, 0)

	for rows.Next() {
		var entry model.LogModel
		err := rows.Scan(
			&entry.ID,
			&entry.RawLog,
			&entry.LogLevel,
			&entry.Source,
			&entry.CreatedAt,
			&entry.RequestID,
			&entry.LoggerName,
		)
		if err != nil {
			return nil, err
		}
		entry.IsDeleted = false
		ret = append(ret, entry)
	}

	return ret, nil

}

func (r *LogReader) durationToPSQLInterval(d *time.Duration) time.Duration {
	if d == nil {
		return time.Second
	}
	return *d
}

func (r *LogReader) GetTimeLineFor(id uint64, before *time.Duration, after *time.Duration) ([]model.LogModel, error) {
	/*
		$1 = id
		$2 = 5 minutes
		$3 = 1 minute
		$4 = ('WARN', 'ERROR', 'CRITICAL', 'FATAL')

	*/
	query := `
			WITH critical_log AS (
				SELECT *
				FROM logs
				WHERE id = $1
			)
			SELECT l.id, l.raw_log, l.log_level, l.source, 
				l.created_at, l.request_id, l.logger_name
			FROM logs l
			CROSS JOIN critical_log cl
			WHERE (
					l.request_id = cl.request_id
					OR (
						l.source = cl.source
						AND l.created_at BETWEEN cl.created_at - $2::interval
												AND cl.created_at + $3::interval
					)
				)
			AND l.log_level = ANY($4)
			AND l.is_deleted = false
			ORDER BY l.created_at ASC
	`

	rows, err := r.tx.Query(query,
		id,
		r.durationToPSQLInterval(before),
		r.durationToPSQLInterval(after),
		pq.Array([]string{"WARN", "ERROR", "CRITICAL", "FATAL"}))

	if err != nil {
		return nil, err
	}

	ret := make([]model.LogModel, 0)

	for rows.Next() {
		var entry model.LogModel
		err := rows.Scan(
			&entry.ID,
			&entry.RawLog,
			&entry.LogLevel,
			&entry.Source,
			&entry.CreatedAt,
			&entry.RequestID,
			&entry.LoggerName,
		)
		if err != nil {
			return nil, err
		}
		entry.IsDeleted = false
		ret = append(ret, entry)
	}

	return ret, nil
}
