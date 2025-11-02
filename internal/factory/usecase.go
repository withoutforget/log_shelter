package factory

import (
	"context"
	"database/sql"

	"log_shelter/internal/usecase"
)

type UsecaseFactory struct {
	ctx            context.Context
	conn           *sql.Conn
	tx             *sql.Tx
	repo_factory   *RepositoryFactory
	reader_factory *ReaderFactory
}

func NewUsecaseFactory(ctx context.Context, conn *sql.Conn,
	tx *sql.Tx,
) *UsecaseFactory {
	return &UsecaseFactory{
		tx: tx, ctx: ctx, conn: conn,
		repo_factory:   NewRepositoryFactory(ctx, tx),
		reader_factory: NewReaderFactory(ctx, tx),
	}
}

func (f *UsecaseFactory) Close() {
	f.conn.Close()
}

func (f *UsecaseFactory) GetAppendLogUsecase() *usecase.AppendLogUsecase {
	return &usecase.AppendLogUsecase{Tx: f.tx, LogRepo: f.repo_factory.GetLogRepository()}
}
