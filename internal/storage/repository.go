package storage

import (
	"context"

	"github.com/Silverman143/character-service/internal/domain/models"
	"github.com/Silverman143/character-service/internal/services/character/dto"
)


type IAppProvider interface {
	App(ctx context.Context, appID int)(models.App, error)
}

type ICharacterProvider interface {
	GetCharacterLevel(ctx context.Context, userID int64) (*int, error)
	CreateCharacter(ctx context.Context, userID int64) error
	GetCharacter(ctx context.Context, userID int64) (*dto.GetCharacterDTO, error)
	GetAllSkins(ctx context.Context) (*dto.GetSkinsDTO, error)
	GetAllLevelPrices(ctx context.Context) (*dto.LevelPriceListDTO, error)
	GetLevelPrice(ctx context.Context, level int16) (*int64, error)
	UpgradeCharacterLevel(ctx context.Context, userID int64) (*int, error)
	ChangeActiveSkin(ctx context.Context, userID int64, skinID int32) error
}