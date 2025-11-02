package factory

import (
	"context"
	"database/sql"

	"log_shelter/internal/infra/reader"
)

type ReaderFactory struct {
	ctx        context.Context
	tx         *sql.Tx
	log_reader *reader.LogReader
}

func NewReaderFactory(ctx context.Context,
	tx *sql.Tx,
) *ReaderFactory {
	return &ReaderFactory{tx: tx, ctx: ctx}
}

func (f *ReaderFactory) GetLogReader() *reader.LogReader {
	if f.log_reader == nil {
		f.log_reader = reader.NewLogReader(f.ctx, f.tx)
	}
	return f.log_reader
}
