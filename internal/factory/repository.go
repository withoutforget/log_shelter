package factory

import (
	"context"
	"database/sql"

	"log_shelter/internal/infra/repository"
)

type RepositoryFactory struct {
	ctx      context.Context
	tx       *sql.Tx
	log_repo *repository.LogRepository
}

func NewRepositoryFactory(ctx context.Context,
	tx *sql.Tx,
) *RepositoryFactory {
	return &RepositoryFactory{tx: tx, ctx: ctx}
}

func (f *RepositoryFactory) GetLogRepository() *repository.LogRepository {
	if f.log_repo == nil {
		f.log_repo = repository.NewLogRepository(f.ctx, f.tx)
	}
	return f.log_repo
}
