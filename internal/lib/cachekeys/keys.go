package cachekeys

import "fmt"

const (
	CharacterLevelPrefix = "character_level:"
	CharacterDataPrefix = "character_data:"

	AllSkinsInfo = "skins_info"
	LevelPrices = "level_prices"
)

// CharacterLevel - return generated key from prefix const and user id
func CharacterLevel(userID int64) string {
    return fmt.Sprintf("%s%d", CharacterLevelPrefix, userID)
}

func CharacterData(userID int64) string {
	return fmt.Sprintf("%s%d", CharacterDataPrefix, userID)
}