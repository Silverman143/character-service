package dto

import characterv1 "github.com/Silverman143/protos_chadnaldo/gen/go/character"

type GetSkinsDTO struct {
	Skins []SkinInfoDTO
}

type SkinInfoDTO struct {
    ID				int			`json:"skin_id" db:"skin_id"`
	Name			string		`json:"character_name" db:"character_name"`
	Lore			string		`json:"character_lore" db:"character_lore"`
	ImageURL 		string		`json:"character_image_url" db:"character_image_url"`
	UnlockLevel		int			`json:"unlock_level" db:"unlock_level"`
	Price			int64		`json:"price" db:"price"`
	RefToBuy 		int			`json:"referrals" db:"referrals"`
	RefToOpen   	int			`json:"referral_to_open" db:"referral_to_open"`
	IsOpened		bool		`json:"is_opened"`
	Stats 			SkinStats   
}

func(s *GetSkinsDTO) UpdateSkinsOpenStatus(currentLevel int) {
    for i := range s.Skins {
        if currentLevel >= s.Skins[i].UnlockLevel {
            s.Skins[i].IsOpened = true
        }
    }
}

func (s *GetSkinsDTO) IsOpened(skinID int) bool {
    for _, skin := range s.Skins {
        if skin.ID == skinID {
            return skin.IsOpened
        }
    }
    return false
}

func (s *GetSkinsDTO) ToGetAllSkinsResponse() *characterv1.GetAllSkinsResponse {
    response := &characterv1.GetAllSkinsResponse{
        Characters: make([]*characterv1.SkinInfo, len(s.Skins)),
    }

    for i, skin := range s.Skins {
        response.Characters[i] = &characterv1.SkinInfo{
            SkinId:          int64(skin.ID),
            ImageUrl:        skin.ImageURL,
            Name:            skin.Name,
            Lore:            skin.Lore,
            Level:           int32(skin.UnlockLevel),
            Price:           skin.Price,
            ReferralsToBuy:  int32(skin.RefToBuy),
            ReferralsToOpen: int32(skin.RefToOpen),
            Bought:          skin.IsOpened, // Предполагаем, что IsOpened соответствует bought
            Stats: &characterv1.SkinStats{
                GamesPlayed: int32(skin.Stats.GamesPlayed),
                HoursPlayed: int32(skin.Stats.HoursPlayed),
                CoinsEarned: skin.Stats.CoinsEarned,
            },
        }
    }

    return response
}

type SkinStats struct {
	GamesPlayed int
	HoursPlayed int
	CoinsEarned int64
}




// // Response with the list of all characters
// message GetAllSkinsResponse {
//     repeated SkinInfo characters = 1;  // List of all characters
// }

// // Information about a character
// message SkinInfo {
//     int64 skin_id = 1;              // ID of the skin
//     string image_url = 2;           // URL of skin image
//     string name = 3;                // Name of the character
//     string lore = 4;                // Skin character lore
//     int32 level = 5;                // Level of the character
//     int32 price = 6;                // Skin price
//     int32 referrals_to_buy = 7;     // Ref to buy
//     int32 referrals_to_open = 8;    // Ref to open without buying
//     bool bought = 9;                // Is user bought this skin
//     SkinStats stats = 10;           // Game stats of the skin

// }

// message SkinStats {
//     int32 games_played = 1;
//     int32 hours_played = 2;
//     int32 coins_earned = 3;
// }