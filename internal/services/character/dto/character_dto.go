package dto


type GetCharacterDTO struct {
	Name 			string		`json:"skin_name" db:"character_name"`
    CurrentLevel  	int    		`json:"current_level" db:"current_level"`
	MiningRate		int64		`json:"mining_rate" db:"mining_force"`
	MiningDuration	int			`json:"mining_duration" db:"mining_duration_minutes"`
    SkinID         	int		 	`json:"current_skin_id" db:"skin_id"`
    SkinImgURL      string 		`json:"skin_image_url" db:"character_image_url"`
}
