package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

type User struct {
	ID                	int64           `json:"id" db:"id"`
	TelegramID        	int64           `json:"telegram_id" db:"telegram_id"`
	TelegramUsername  	sql.NullString  `json:"telegram_username" db:"telegram_username"`
	TelegramFirstName 	sql.NullString  `json:"telegram_first_name" db:"telegram_first_name"`
	TelegramLastName  	sql.NullString  `json:"telegram_last_name" db:"telegram_last_name"`
	TelegramAvatarURL 	sql.NullString  `json:"telegram_avatar_url" db:"telegram_avatar_url"`
	LanguageCode      	sql.NullString  `json:"language_code" db:"language_code"`
	IsTelegramPremium	bool            `json:"is_premium" db:"is_premium"`
	CreatedAt         	time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt         	time.Time       `json:"updated_at" db:"updated_at"`
	LastActiveAt      	sql.NullTime    `json:"last_active_at" db:"last_active_at"`
	IsBlocked         	bool            `json:"is_blocked" db:"is_blocked"`
	Settings          	json.RawMessage `json:"settings" db:"settings"`
	Metadata          	json.RawMessage `json:"metadata" db:"metadata"`
}

