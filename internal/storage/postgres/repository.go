package postgres

import (
	"github.com/Silverman143/character-service/internal/storage"
)

type Repository struct {
    storage.IAppProvider
    storage.ICharacterProvider
}

func NewRepository(st *Storage) *Repository {
    return &Repository{
        IAppProvider:  NewAppProvider(st),
        ICharacterProvider: NewCharacterProvider(st),
    }
}