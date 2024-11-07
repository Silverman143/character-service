package characterservice

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	referralgrpc "github.com/Silverman143/character-service/internal/clients/referral/grpc"
	usergrpc "github.com/Silverman143/character-service/internal/clients/user/grpc"
	kafkaproducer "github.com/Silverman143/character-service/internal/kafka/producer"
	"github.com/Silverman143/character-service/internal/lib/cachekeys"
	cache "github.com/Silverman143/character-service/internal/redis"
	"github.com/Silverman143/character-service/internal/services/character/dto"
	"github.com/Silverman143/character-service/internal/storage"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

type Character struct {
	log *slog.Logger
	//implementa stofage interfaces
	appProvider storage.IAppProvider
	characterProvider storage.ICharacterProvider
    cache *cache.RedisCache
    kafkaProducer *kafkaproducer.KafkaProducer
	userClient *usergrpc.Client
	referralClient *referralgrpc.Client

}


func New(	log * slog.Logger, 
			appProvider storage.IAppProvider, 
			characterProvider storage.ICharacterProvider,
			cache *cache.RedisCache, 
			kafkaProducer *kafkaproducer.KafkaProducer, 
			userClient *usergrpc.Client,
			referralClient *referralgrpc.Client) *Character{
	return &Character{
		log: 					log,
		appProvider: 			appProvider,
		characterProvider: 		characterProvider,	
        cache:                  cache,
        kafkaProducer:          kafkaProducer,
		userClient: userClient,
		referralClient:  		referralClient,
	}
}

func (c *Character) CreateCharacter(ctx context.Context, userID int64) error {
	const op = "services.character.CreateCharacter"
	logger := c.log.With("op", op)

	logger.Info("try to create character", "userID", userID)

	if err := c.characterProvider.CreateCharacter(ctx, userID); err != nil {
		logger.Error("Error with createing character", "userID", userID, "error", err)
		return err
	}

	logger.Info("character created successfully", "userID", userID)
	return nil
}

// GetCharacterLevel - returns current level of users character
func (c *Character) GetCharacterLevel(ctx context.Context, userID int64)(*int, error){
	const op = "services.character.GetCharacterLevel"
	logger := c.log.With("op", op)

	logger.Info("try to get character level", "userID", userID)

	var level *int

	// Check cached data
    levelCacheKey := cachekeys.CharacterLevel(userID)
    level, err := c.cache.GetInt(ctx, levelCacheKey)

    if err == nil {
        logger.Info("Level found in cache, returning cached level")
        return level, nil
    }
    if err != redis.Nil {
        logger.Error("Error accessing Redis cache", "error", err)
    }

	level, err = c.characterProvider.GetCharacterLevel(ctx, userID)
	if err != nil{
		logger.Error("Error with getting character level", "userID", userID, "error", err)
		return level, fmt.Errorf("%s:%w", op, err)
	}

	err = c.cache.SetInt(ctx, levelCacheKey, *level, c.cache.Lifetime)
	if err != nil {
		logger.Error("failed to cache user character level", "error", err)
	}

	logger.Info("character level getted successfully", "userID", userID)

	return level, nil
}

// GetCharacter - returns current user character data
func (c *Character) GetCharacter(ctx context.Context, userID int64)(*dto.GetCharacterDTO, error){
	const op = "services.character.GetCharacter"
	logger := c.log.With("op", op)

	logger.Info("try to get character", "userID", userID)

	var characterDto *dto.GetCharacterDTO

	// Get from cache
    characterCacheKey := cachekeys.CharacterData(userID)
	err := c.cache.Get(ctx, characterCacheKey, characterDto)

	if err == nil{
		return characterDto, nil
	}

	if !errors.Is(err, redis.Nil){
		logger.Error("error with getting cached character", "error", err)
	}
	// Get from db
	characterDto, err = c.characterProvider.GetCharacter(ctx, userID)

	if err != nil {
		logger.Error("Error with getting character", "userID", userID, "error", err)
		return &dto.GetCharacterDTO{}, fmt.Errorf("%s:%w", op, err)
	}
	// Save to cache
	err = c.cache.Set(ctx, characterCacheKey, characterDto, c.cache.Lifetime)

	if err != nil{
		logger.Error("error with saving character in cache", "error", err)
	}

	return characterDto, nil
}

// GetSkins - get all skins data
func (c *Character) GetSkins(ctx context.Context, userID int64)(*dto.GetSkinsDTO, error){
	const op = "service.character.GetSkins"
	logger := c.log.With("op", op)


    var (
        level    *int
        skinsDTO *dto.GetSkinsDTO
        levelErr, skinsErr error
    )

    // Параллельное получение уровня и скинов
    group, ctx := errgroup.WithContext(ctx)

    group.Go(func() error {
        level, levelErr = c.GetCharacterLevel(ctx, userID)
        return levelErr
    })

    group.Go(func() error {
        // Получение скинов из кэша
        err := c.cache.Get(ctx, cachekeys.AllSkinsInfo, &skinsDTO)
        if err == nil {
            return nil
        }
        if err != redis.Nil {
            logger.Error("Error getting skins from cache", "error", err)
        }

        // Если данных нет в кэше, получаем из базы данных
        skinsDTO, skinsErr = c.characterProvider.GetAllSkins(ctx)
        if skinsErr == nil {
            // Кэшируем результат
            cacheErr := c.cache.Set(ctx, cachekeys.AllSkinsInfo, skinsDTO, c.cache.Lifetime)
            if cacheErr != nil {
                logger.Error("Failed to cache skins", "error", cacheErr)
            }
        }
        return skinsErr
    })

    // Ожидаем завершения всех горутин
    if err := group.Wait(); err != nil {
        return nil, fmt.Errorf("%s: %w", op, err)
    }

    // Обновляем статус открытия скинов
    skinsDTO.UpdateSkinsOpenStatus(*level)

	return skinsDTO, nil
}


func (c *Character) GetLevelsPrices(ctx context.Context) (*dto.LevelPriceListDTO, error) {
	const op = "services.character.CacheLevelPrices"
	logger := c.log.With("op", op)

	// Создаем карту для хранения цен уровней
	var pricesList dto.LevelPriceListDTO

	// Пытаемся получить данные из кэша
	err := c.cache.Get(ctx, cachekeys.LevelPrices, &pricesList)
	if err == nil {
		// Если данные найдены в кэше, возвращаем их
		return &pricesList, nil
	}

	// Если ошибка - это отсутствие данных в кэше
	if err == redis.Nil {
		logger.Info("levels prices not exist in cache")
	} else {
		// Логируем любую другую ошибку кэширования
		logger.Error("%s: error getting levels prices from cache", "error", err)
	}

	// Получаем цены уровней из characterProvider (Postgres)
	levels, err := c.characterProvider.GetAllLevelPrices(ctx)
	if err != nil {
		logger.Error("%s: error getting levels prices from postgres", "error", err)
		return nil, fmt.Errorf("%s: failed to get level prices: %w", op, err)
	}

	// Кэшируем данные на 24 часа
	if err := c.cache.Set(ctx, cachekeys.LevelPrices, levels, 24*time.Hour); err != nil {
		logger.Error("%s: error caching levels prices", "error", err)
	}

	// Возвращаем карту цен уровней
	return levels, nil
}

func (c *Character) ChangeActiveSkin(ctx context.Context, userID int64, skinID int32) error {
    const op = "services.character.ChangeActiveSkin"
    logger := c.log.With("op", op)

    skins, err := c.GetSkins(ctx, userID)
    if err != nil {
        logger.Error("Error getting user skins", "userID", userID, "error", err)
        return fmt.Errorf("%s: %w", op, err)
    }

    if !skins.IsOpened(int(skinID)) {
        return ErrSkinIsNotOpened
    }

    if err := c.characterProvider.ChangeActiveSkin(ctx, userID, skinID); err != nil {
        logger.Error("Error changing active skin", "userID", userID, "skinID", skinID, "error", err)
        return fmt.Errorf("%s: %w", op, err)
    }

    logger.Info("Active skin changed successfully", "userID", userID, "skinID", skinID)
    return nil
}

