package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Silverman143/character-service/internal/services/character/dto"
	"github.com/doug-martin/goqu/v9"
)

type PostgresCharacterProvider struct {
	storage *Storage
}

func NewCharacterProvider(storage *Storage) *PostgresCharacterProvider{
	return &PostgresCharacterProvider{
		storage: storage,
	}
}

func (s *PostgresCharacterProvider) CreateCharacter(ctx context.Context, userID int64) error {
	const op = "storage.postgres.CreateCharacter"
	dialect := goqu.Dialect("postgres")

	insertQuery := dialect.Insert(TableCharacters).
		Cols("user_id").
		Vals(goqu.Vals{userID}).
		OnConflict(goqu.DoNothing())

	query, args, err := insertQuery.ToSQL()
	if err != nil {
		return fmt.Errorf("%s: failed to build query: %w", op, err)
	}

	_, err = s.storage.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: failed to execute query: %w", op, err)
	}

	return nil
}

func (s *PostgresCharacterProvider) GetCharacterLevel(ctx context.Context, userID int64) (*int, error){
	const op = "storage.postgres.getCharacterLevel"
	dialect := goqu.Dialect("postgres")

	var level int
    selectQuery := dialect.From(TableCharacters).
        Select(
			"current_level",
        ).
        Where(goqu.C("user_id").Eq(userID))

    query, args, err := selectQuery.ToSQL()

    if err != nil {
        return &level, fmt.Errorf("%s: failed to build query: %w", op, err)
    }

    err = s.storage.db.GetContext(ctx, &level, query, args...)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return &level, fmt.Errorf("%s: %w", op, ErrCharacterNotFound)
        }
        return &level, fmt.Errorf("%s: failed to execute query: %w", op, err)
	}
	return &level, nil
}

func (s *PostgresCharacterProvider) GetCharacter(ctx context.Context, userID int64) (*dto.GetCharacterDTO, error) {
    const op = "storage.postgres.getCharacter"

    dialect := goqu.Dialect("postgres")
    selectQuery := dialect.From(TableCharacters).
        LeftJoin(
            goqu.T("character_skins"),
            goqu.On(goqu.Ex{"characters.current_skin_id": goqu.I("character_skins.skin_id")}),
        ).
        LeftJoin(
            goqu.T("character_levels"),
            goqu.On(goqu.Ex{"characters.current_level": goqu.I("character_levels.level_number")}),
        ).
        Where(goqu.Ex{"characters.user_id": userID}).
        Select(
            "characters.current_level",
            "character_skins.skin_id",
            "character_skins.character_name",
            "character_skins.character_image_url",
			"character_levels.mining_force",
			"character_levels.mining_duration_minutes",
        )

    query, args, err := selectQuery.ToSQL()
    if err != nil {
        return nil, fmt.Errorf("%s: failed to build query: %w", op, err)
    }

    var character dto.GetCharacterDTO
    err = s.storage.db.GetContext(ctx, &character, query, args...)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, fmt.Errorf("%s: %w", op, ErrCharacterNotFound)
        }
        return nil, fmt.Errorf("%s: failed to execute query: %w", op, err)
    }

    return &character, nil
}

func (s *PostgresCharacterProvider) GetAllSkins(ctx context.Context) (*dto.GetSkinsDTO, error) {
	const op = "storage.postgres.getCharacter"

	var skins dto.GetSkinsDTO
	dialect := goqu.Dialect("postgres")
    
    query := dialect.From(TableCharacterSkins).
        LeftJoin(
            goqu.T("character_levels"),
            goqu.On(goqu.Ex{"character_skins.unlock_level": goqu.I("character_levels.level_number")}),
        ).
        Select(
            "character_skins.skin_id",
            "character_skins.character_name",
            "character_skins.character_lore",
            "character_skins.character_image_url",
            "character_skins.unlock_level",
            "character_levels.price",
            "character_levels.referrals",
            "character_levels.referral_to_open",
        )

    sql, args, err := query.ToSQL()
    if err != nil {
        return &skins, fmt.Errorf("%s: failed to build query: %w", op, err)
    }

    rows, err := s.storage.db.Query(sql, args...)
    if err != nil {
        return &skins, fmt.Errorf("%s: failed to executes a query: %w", op, err)
    }
    defer rows.Close()

    for rows.Next() {
        var skin dto.SkinInfoDTO
        err := rows.Scan(
            &skin.ID,
            &skin.Name,
            &skin.Lore,
            &skin.ImageURL,
            &skin.UnlockLevel,
            &skin.Price,
            &skin.RefToBuy,
            &skin.RefToOpen,
        )
        if err != nil {
            return &skins, err
        }
        
        // IsOpened и Stats не хранятся в базе данных, поэтому устанавливаем значения по умолчанию
        skin.IsOpened = false
        skin.Stats = dto.SkinStats{
            GamesPlayed: 0,
            HoursPlayed: 0,
            CoinsEarned: 0,
        }

        skins.Skins = append(skins.Skins, skin)
    }

    if err = rows.Err(); err != nil {
        return &skins, err
    }

	return &skins, nil
}

func (s *PostgresCharacterProvider) GetAllLevelPrices(ctx context.Context) (*dto.LevelPriceListDTO, error) {
    const op = "storage.postgres.GetAllLevelPrices"

	dialect := goqu.Dialect("postgres")

    query := dialect.From(TableCharacterLevels).
        Select("level_number", "price", "referrals", "referral_to_open").
        Order(goqu.I("level_number").Asc())

    sql, args, err := query.ToSQL()
    if err != nil {
        return nil, fmt.Errorf("%s: failed to build query: %w", op, err)
    }

    var levels []dto.LevelPriceDTO
    err = s.storage.db.SelectContext(ctx, &levels, sql, args...)
    if err != nil {
        return nil, fmt.Errorf("%s: failed to execute query: %w", op, err)
    }

    return &dto.LevelPriceListDTO{Skins: levels}, nil

}

func (s *PostgresCharacterProvider) GetLevelPrice(ctx context.Context, level int16) (*int64, error) {
    const op = "storage.postgres.GetLevelPrice"

	dialect := goqu.Dialect("postgres")

    selectQuery := dialect.From(TableCharacterLevels).
        Select("price"). Where(goqu.C("level_number").Eq(level))
	
    query, args, err := selectQuery.ToSQL()

	var price *int64

    if err != nil {
        return price, fmt.Errorf("%s: failed to build query: %w", op, err)
    }

    err = s.storage.db.GetContext(ctx, &level, query, args...)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return price, fmt.Errorf("%s: %w", op, ErrCharacterNotFound)
        }
        return price, fmt.Errorf("%s: failed to execute query: %w", op, err)
	}
	return price, nil
}

func (s *PostgresCharacterProvider) UpgradeCharacterLevel(ctx context.Context, userID int64) (*int, error) {
	const op = "storage.postgres.UpgradeCharacter"

	dialect := goqu.Dialect("postgres")

	// Создаем запрос для увеличения уровня на 1 с использованием RETURNING для возврата нового уровня
	updateQuery := dialect.Update(TableCharacters).
		Set(goqu.Record{"current_level": goqu.L("current_level + 1")}).
		Where(goqu.C("user_id").Eq(userID)).
		Returning("current_level") // Добавляем RETURNING, чтобы сразу вернуть обновленное значение

	// Преобразуем запрос в SQL
	sql, args, err := updateQuery.ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to build query: %w", op, err)
	}

	// Переменная для хранения нового уровня
	var currentLevel int

	// Выполняем SQL-запрос, который сразу вернет новый уровень
	err = s.storage.db.GetContext(ctx, &currentLevel, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to upgrade and retrieve current level: %w", op, err)
	}

	return &currentLevel, nil
}

func (s *PostgresCharacterProvider) ChangeActiveSkin(ctx context.Context, userID int64, skinID int32) error {
    const op = "storage.postgres.SelectActiveSkin"

    dialect := goqu.Dialect("postgres")

    // Создаем запрос для обновления current_skin_id
    updateQuery := dialect.Update(TableCharacters).
        Set(goqu.Record{"current_skin_id": skinID}).
        Where(goqu.C("user_id").Eq(userID))

    // Преобразуем запрос в SQL
    sql, args, err := updateQuery.ToSQL()
    if err != nil {
        return fmt.Errorf("%s: failed to build query: %w", op, err)
    }

    // Выполняем запрос
    _, err = s.storage.db.ExecContext(ctx, sql, args...)
    if err != nil {
        return fmt.Errorf("%s: failed to execute query: %w", op, err)
    }

    return nil
}
