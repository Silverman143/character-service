package app

import (
	"log/slog"

	grpcapp "github.com/Silverman143/character-service/internal/app/grpc"
	referralgrpc "github.com/Silverman143/character-service/internal/clients/referral/grpc"
	usergrpc "github.com/Silverman143/character-service/internal/clients/user/grpc"
	"github.com/Silverman143/character-service/internal/config"
	kafkaproducer "github.com/Silverman143/character-service/internal/kafka/producer"
	cache "github.com/Silverman143/character-service/internal/redis"
	characterService "github.com/Silverman143/character-service/internal/services/character"
	"github.com/Silverman143/character-service/internal/storage/postgres"
)

type App struct {
	GRPCServer *grpcapp.App
	KafkaProducer *kafkaproducer.KafkaProducer

}

func New (	log *slog.Logger, 
			config *config.Config, 
			storage *postgres.Storage, 
			cache *cache.RedisCache, 
			kafkaProducer *kafkaproducer.KafkaProducer,
			userClient *usergrpc.Client, 
			referralClient *referralgrpc.Client  ) *App{

	repo := postgres.NewRepository(storage)

	characterService := characterService.New(log, repo, repo, cache, kafkaProducer, userClient, referralClient)

	gRPCApp := grpcapp.New(log, characterService, config.GRPC.Port)

	return &App{
		GRPCServer: gRPCApp,
		KafkaProducer: kafkaProducer,
	}
}
