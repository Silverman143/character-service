package character

import (
	characterservice "github.com/Silverman143/character-service/internal/services/character"
	"github.com/Silverman143/character-service/internal/services/character/dto"
	characterv1 "github.com/Silverman143/protos_chadnaldo/gen/go/character"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Character interface {
	GetCharacterLevel(ctx context.Context, user_id int64)(*int, error)
	CreateCharacter(ctx context.Context, user_id int64) error
	GetCharacter(ctx context.Context, user_id int64)(*dto.GetCharacterDTO, error)
	GetSkins(ctx context.Context, user_id int64)(*dto.GetSkinsDTO, error)
	LevelUpCharacter(ctx context.Context, userID int64) (newLevel *int, coinsBalance *int64, err error)
	ChangeActiveSkin(ctx context.Context, userID int64, skinID int32 ) error
}

type serverAPI struct {
	characterv1.UnimplementedCharacterServer
	character Character
}

func Register(gRPCServer *grpc.Server, 	character *characterservice.Character) {
	characterv1.RegisterCharacterServer(gRPCServer, &serverAPI{character: character})
}

const (
	emptyValue = ""
	emptyInt = 0
)

func (s *serverAPI) GetCharacterLevel (ctx context.Context, req *characterv1.GetCharacterLevelRequest) (*characterv1.GetCharacterLevelResponse, error ){
	if req.GetUserId() == emptyInt{
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	level, err := s.character.GetCharacterLevel(ctx, req.UserId);
	if err != nil{
		return &characterv1.GetCharacterLevelResponse{
			Level: 0,
		}, status.Error(codes.Internal, "character is not exists")
	}

	return &characterv1.GetCharacterLevelResponse{
		Level: int32(*level),
	}, nil
}

func (s *serverAPI) CreateCharacter (ctx context.Context, req *characterv1.CreateCharacterRequest) (*characterv1.CreateCharacterResponse, error ){
	if req.GetUserId() == emptyInt{
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	if err := s.character.CreateCharacter(ctx, req.UserId); err != nil{
		return &characterv1.CreateCharacterResponse{
			Success: false,
		}, status.Error(codes.Internal, "could not create character")
	}
	return &characterv1.CreateCharacterResponse{
		Success: true,
	}, nil
}

func (s *serverAPI) GetCharacter (ctx context.Context, req *characterv1.GetCharacterRequest) (*characterv1.GetCharacterResponse, error ){
	if req.GetUserId() == emptyInt{
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	characterDto, err := s.character.GetCharacter(ctx, req.UserId)

	if err != nil{
		return &characterv1.GetCharacterResponse{}, status.Error(codes.Internal, "could not get character")
	}

	return &characterv1.GetCharacterResponse{
		Name: characterDto.Name,
		Level: int32(characterDto.CurrentLevel),
		MiningRate: characterDto.MiningRate,
		MiningDuration: int32(characterDto.MiningDuration),
		CurrentSkinId: int32(characterDto.SkinID),
		CurrentSkinImageUrl: characterDto.SkinImgURL,
	}, nil
}

func (s *serverAPI) GetAllSkins (ctx context.Context, req *characterv1.GetAllSkinsRequest) (*characterv1.GetAllSkinsResponse, error ){
	if req.GetUserId() == emptyInt{
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	skinsDTO, err := s.character.GetSkins(ctx, req.UserId)

	if err != nil{
		return &characterv1.GetAllSkinsResponse{}, status.Error(codes.Internal, "could not get character")
	}

	return skinsDTO.ToGetAllSkinsResponse(), nil
}

func (s *serverAPI) LevelUpCharacter (ctx context.Context, req *characterv1.LevelUpCharacterRequest) (*characterv1.LevelUpCharacterResponse, error){
	if req.GetUserId() == emptyInt{
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	level, balance, err := s.character.LevelUpCharacter(ctx, req.UserId)
	
	if err != nil{
		return &characterv1.LevelUpCharacterResponse{Success: false}, status.Error(codes.Internal, "could not upgrade character level")
	}
	return &characterv1.LevelUpCharacterResponse{Success: true, NewLevel: int32(*level), CoinsBalance: *balance }, nil
}


func (s *serverAPI) SelectActiveSkin (ctx context.Context, req *characterv1.SelectActiveSkinRequest) (*characterv1.SelectActiveSkinResponse, error){
	if req.GetUserId() == emptyInt{
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}
	if req.GetSkinId() == emptyInt{
		return nil, status.Error(codes.InvalidArgument, "skin id is required")
	}

	if err := s.character.ChangeActiveSkin(ctx, req.UserId, req.SkinId); err != nil{
		return &characterv1.SelectActiveSkinResponse{Success: false, Message: "error with changing skin"}, err 
	}

	return &characterv1.SelectActiveSkinResponse{Success: true}, nil 
}
