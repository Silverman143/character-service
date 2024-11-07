package postgres

import (
	"context"
	"fmt"

	"github.com/Silverman143/character-service/internal/domain/models"
)

type PostgresAppProvider struct {
	storage *Storage
}

func NewAppProvider(storage *Storage) *PostgresAppProvider{
	return &PostgresAppProvider{
		storage: storage,
	}
}

func (s *PostgresAppProvider) App(ctx context.Context, appID int) (models.App, error){
	const op = "storage.postgres.App"
	fmt.Printf("App provider not implement %s", op)
	return models.App{}, nil
}