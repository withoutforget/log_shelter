package factory

import (
	"context"
	"log_shelter/internal/infra"
)

type Factory struct {
	postgres_infra *infra.PostgresInfra
}

func NewFactory(postgres_infra *infra.PostgresInfra) *Factory {
	return &Factory{
		postgres_infra: postgres_infra,
	}
}

func (f *Factory) GetUsecaseFactory(ctx context.Context) (*UsecaseFactory, error) {
	conn, tx, err := f.postgres_infra.GetTranscation()
	if err != nil {
		return nil, err
	}
	return NewUsecaseFactory(ctx, conn, tx), nil
}
