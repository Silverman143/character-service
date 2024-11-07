package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/Silverman143/character-service/internal/config"
	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
    client *redis.Client
    Lifetime time.Duration
	logger *slog.Logger
}

func NewRedisCache(cfg config.RedisConfig, log *slog.Logger) (*RedisCache, error) {
	const op = "redis.NewRedisCache"
	logger := log.With("op", op)

    client := redis.NewClient(&redis.Options{
        Addr: cfg.Addr,
		Password: cfg.Password,
		DB: cfg.DB,
    })

    // Проверка подключения
    _, err := client.Ping(context.Background()).Result()
    if err != nil {
		logger.Warn("Couldn't connect Redis")
        return nil, fmt.Errorf("%s:%w", op, err)
    }
    return &RedisCache{client: client, logger: log, Lifetime: cfg.Lifetime}, nil
}

func (r *RedisCache) SetString(ctx context.Context, key string, value string, expiration time.Duration) error {
	const op = "redis.setString"
	logger := r.logger.With("op", op)
	if err := r.client.Set(ctx, key, value, expiration).Err(); err != nil {
		logger.Error("couldn't set value", "error", err)
        return fmt.Errorf("%s:%w", op, err)
	}
    return nil
}

func (r *RedisCache) GetString(ctx context.Context, key string) (*string, error) {
    const op = "redis.getString"
    logger := r.logger.With("op", op)

    val, err := r.client.Get(ctx, key).Result()
    if err != nil {
        if err == redis.Nil {
            logger.Debug("key not found", "key", key)
            return &val, fmt.Errorf("%s: key not found: %w", op, err)
        }
        logger.Error("couldn't get value", "error", err)
        return &val, fmt.Errorf("%s: %w", op, err)
    }

    return &val, nil
}

func (r *RedisCache) GetInt(ctx context.Context, key string) (*int, error) {
    const op = "redis.getInt"
    logger := r.logger.With("op", op)

    val, err := r.client.Get(ctx, key).Int()
    if err != nil {
        if err == redis.Nil {
            logger.Debug("key not found", "key", key)
            return &val, fmt.Errorf("%s: key not found: %w", op, err)
        }
        logger.Error("couldn't get value", "error", err)
        return &val, fmt.Errorf("%s: %w", op, err)
    }

    return &val, nil
}

func (r *RedisCache) SetInt(ctx context.Context, key string, value int, expiration time.Duration) error {
    const op = "redis.setInt"
    logger := r.logger.With("op", op)

    err := r.client.Set(ctx, key, value, expiration).Err()
    if err != nil {
        logger.Error("couldn't set value", "error", err)
        return fmt.Errorf("%s: %w", op, err)
    }

    return nil
}


// Set сохраняет любой объект в Redis
func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
    const op = "redis.set"
    logger := r.logger.With("op", op)

    jsonData, err := json.Marshal(value)
    if err != nil {
        logger.Error("couldn't json marshal value", "error", err)
        return fmt.Errorf("%s: %w", op, err)
    }

    err = r.client.Set(ctx, key, jsonData, expiration).Err()
    if err != nil {
        logger.Error("couldn't set value", "error", err)
        return fmt.Errorf("%s: %w", op, err)
    }
    return nil
}

// Get получает объект из Redis и десериализует его в указанный тип
func (r *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
    const op = "redis.get"
    logger := r.logger.With("op", op)

    cachedData, err := r.client.Get(ctx, key).Result()
    if err != nil {
        if err == redis.Nil {
            logger.Debug("key not found", "key", key)
            return fmt.Errorf("%s: key not found: %w", op, err)
        }
        logger.Error("couldn't get value", "error", err)
        return fmt.Errorf("%s: %w", op, err)
    }

    err = json.Unmarshal([]byte(cachedData), dest)
    if err != nil {
        logger.Error("couldn't json unmarshal value", "error", err)
        return fmt.Errorf("%s: %w", op, err)
    }

    return nil
}


// Exists проверяет наличие ключа в кэше
func (r *RedisCache) Exists(ctx context.Context, key string) (*int64, error) {
    const op = "redis.Exists"
    level, err := r.client.Exists(ctx, key).Result()
    if err != nil {
        return &level, fmt.Errorf("%s:%w", op, err)
    }
    return &level, nil
}

// Delete удаляет ключ из кэша
func (r *RedisCache) Delete(ctx context.Context, key string) error {
    const op = "redis.delete"
	logger := r.logger.With("op", op)

	if err:= r.client.Del(ctx, key).Err(); err != nil{
		logger.Error("couldn't delete value", "error", err)
		return fmt.Errorf("%s:%w", op, err)
	}
	return nil
}

// Close закрывает соединение с Redis
func (r *RedisCache) Close() error {
	const op = "redis.close"
	logger := r.logger.With("op", op)

	if err := r.client.Close(); err != nil {
		logger.Error("couldn't close Redis client", "error", err)
        return fmt.Errorf("%s:%w", op, err)
	}
	logger.Info("Redis client closed successfully")
    return nil
}