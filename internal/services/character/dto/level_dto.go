package dto

// SkinPrice представляет скин и его цены
type LevelPriceDTO struct {
    Level                 int     `json:"level_number" db:"level_number"`
    CoinsPrice            int64   `json:"coins_price" db:"price"`
    ReferralsPrice        int64  `json:"referrals" db:"referrals"`
    ReferralsForFreeOpen  int64  `json:"referral_to_open" db:"referral_to_open"`
}

// SkinPriceList представляет список всех скинов и их цен
type LevelPriceListDTO struct {
    Skins []LevelPriceDTO `json:"skins"`
}

// GetSkinPrice возвращает цены для конкретного скина по имени
func (spl *LevelPriceListDTO) GetLevelPrice(level int) (LevelPriceDTO, bool) {
    for _, skin := range spl.Skins {
        if skin.Level == level {
            return skin, true
        }
    }
    return LevelPriceDTO{}, false
}