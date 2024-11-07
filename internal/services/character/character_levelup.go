package characterservice

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Silverman143/character-service/internal/lib/cachekeys"
	"github.com/Silverman143/character-service/internal/services/character/dto"
	"github.com/google/uuid"
)

func (c *Character) LevelUpCharacter(ctx context.Context, userID int64) (newLevel *int, coinsBalance *int64, err error) {
	const op = "service.character.LevelUpCharacter"
	logger := c.log.With("op", op)

	// Получаем текущий уровень персонажа
	level, err := c.GetCharacterLevel(ctx, userID)
	if err != nil {
		logger.Error("failed to get character level")
		return nil, nil, fmt.Errorf("%s: failed to get character level: %w", op, err)
	}

	// Получаем цены уровней
	levelsPrices, err := c.GetLevelsPrices(ctx)
	if err != nil {
		logger.Error("failed to get levels prices")
		return level, nil, fmt.Errorf("%s: failed to get levels prices: %w", op, err)
	}

	// Проверяем, есть ли цена для следующего уровня
	nextLevelPrice, exists := levelsPrices.GetLevelPrice(*level + 1)
	if !exists {
		logger.Error("no price for the next level")
		return level, nil, fmt.Errorf("%s: no price for the next level", op)
	}

	// Получаем количество монет и рефералов пользователя
	coins, referrals, err := c.getUserInfo(ctx, userID)
	if err != nil {
		logger.Error("Error with getting user info", slog.Any("err", err))
		return level, nil, fmt.Errorf("%s: %w", op, err)
	}

	// Проверяем возможность повышения уровня
	canLevelUp, isFreeLevelUp := c.canLevelUp(nextLevelPrice, coins, referrals)
	if !canLevelUp {
		logger.Error("not enough coins or referrals to upgrade level")
		return level, nil, fmt.Errorf("%s: not enough coins or referrals to upgrade level", op)
	}

	// Повышаем уровень
	newLevel, coins, err = c.upgradeLevel(ctx, userID, isFreeLevelUp, coins, nextLevelPrice.CoinsPrice)
	if err != nil {
		logger.Error("Error wuth upgrade level", slog.Any("error", err))
		return level, nil, fmt.Errorf("%s: %w", op, err)
	}

	// Обновляем кэш нового уровня
	if err := c.cacheNewLevel(ctx, userID, *newLevel); err != nil {
		logger.Error("failed to cache user character level", "error", err)
	}

	return newLevel, &coins, nil
}

func (c *Character) getUserInfo(ctx context.Context, userID int64) (coins int64, referrals int, err error) {
	coins, err = c.userClient.GetCoinsAmount(ctx, userID)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get user coins balance: %w", err)
	}

	referrals, err = c.referralClient.GetReferralsAmount(ctx, userID)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get referrals amount: %w", err)
	}

	return coins, referrals, nil
}

func (c *Character) canLevelUp(price dto.LevelPriceDTO, coins int64, referrals int) (bool, bool) {
	if price.CoinsPrice <= coins {
		return true, false
	}
	if price.ReferralsForFreeOpen <= int64(referrals) {
		return true, true
	}
	return false, false
}

func (c *Character) upgradeLevel(ctx context.Context, userID int64, isFreeLevelUp bool, coins int64, price int64) (*int, int64, error) {
	var newLevel *int
	var err error

	if !isFreeLevelUp {
		paymentID := uuid.New().String()
		if err := c.userClient.InitiatePayment(ctx, userID, price, paymentID); err != nil {
			return newLevel, coins, fmt.Errorf("failed to initiate payment: %w", err)
		}

		newLevel, err = c.characterProvider.UpgradeCharacterLevel(ctx, userID)
		if err != nil {
			if rollbackErr := c.userClient.FinalizePayment(ctx, paymentID, false); rollbackErr != nil {
				c.log.Error("Error rolling back payment", "userID", userID, "paymentID", paymentID, "error", rollbackErr)
			}
			return newLevel, coins, fmt.Errorf("failed to upgrade character level: %w", err)
		}

		if err := c.userClient.FinalizePayment(ctx, paymentID, true); err != nil {
			return newLevel, coins, fmt.Errorf("failed to finalize payment: %w", err)
		}

		coins -= price
	} else {
		newLevel, err = c.characterProvider.UpgradeCharacterLevel(ctx, userID)
		if err != nil {
			return newLevel, coins, fmt.Errorf("failed to upgrade character level: %w", err)
		}
	}

	return newLevel, coins, nil
}

func (c *Character) cacheNewLevel(ctx context.Context, userID int64, newLevel int) error {
	cachekey := cachekeys.CharacterLevel(userID)
	return c.cache.SetInt(ctx, cachekey, newLevel, c.cache.Lifetime)
}