package postgres

import (
	"fmt"

	"github.com/Silverman143/character-service/internal/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // Явно импортируем драйвер PostgreSQL
)

type Storage struct {
	db *sqlx.DB
}

func New(cfg *config.PgSql) (*Storage, error) {
	const op = "storage.postgres.postgres.NewDBConnection"

	connStr := fmt.Sprintf("user=%s dbname=%s sslmode=%s password=%s host=%s",
        cfg.User, cfg.DbName, cfg.SSLMode, cfg.Password, cfg.Host)

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Storage{db: db}, nil
}

func (s *Storage) Stop() error {
	return s.db.Close()
}




