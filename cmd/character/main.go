package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Silverman143/character-service/internal/app"
	referralgrpc "github.com/Silverman143/character-service/internal/clients/referral/grpc"
	usergrpc "github.com/Silverman143/character-service/internal/clients/user/grpc"
	"github.com/Silverman143/character-service/internal/config"
	kafkaconsumer "github.com/Silverman143/character-service/internal/kafka/consumer"
	kafkaproducer "github.com/Silverman143/character-service/internal/kafka/producer"
	slogpretty "github.com/Silverman143/character-service/internal/lib/cachekeys/logger/pretter"
	cache "github.com/Silverman143/character-service/internal/redis"
	"github.com/Silverman143/character-service/internal/storage/postgres"
)

const(
	envLocal = "local"
	envDev = "dev"
	envProd = "prod"
)

func main(){
    cfg := config.MustLoad()
    log := setupLogger(cfg.Env)

    if cfg.Env == envLocal {
        log.Info("starting application", slog.Any("cfg", cfg))
    }

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    storage, err := postgres.New(&cfg.PgSql)
    if err != nil {
        log.Error("Failed to connect to postgres", slog.String("error", err.Error()))
        os.Exit(1)
    }
    defer storage.Stop()


    cache, err := cache.NewRedisCache(cfg.Redis, log)
    if err != nil {
        log.Error("Failed to connect to Redis", slog.String("error", err.Error()))
        os.Exit(1)
    }
    defer cache.Close()

    kafkaProducer, err := kafkaproducer.NewKafkaProducer(cfg.Kafka, log)
    if err != nil {
        log.Error("Failed to create kafka producer", slog.String("error", err.Error()))
        os.Exit(1)
    }
    defer kafkaProducer.Close()

    kafkaConsumer, err := kafkaconsumer.NewKafkaConsumer(cfg.Kafka, log)
    if err != nil{
        log.Error("Failed to create kafka consumer", slog.String("error", err.Error()))
        os.Exit(1)
    }

	userClient, err := usergrpc.New(ctx, log, &cfg.Clients.User)
	if err != nil{
		log.Error("Failed to connect to UserClient", slog.String("error", err.Error()))
        os.Exit(1)
	}

	referralClient, err := referralgrpc.New(ctx, log, &cfg.Clients.Referral)
	if err != nil{
		log.Error("Failed to connect to Referral Service", slog.String("error", err.Error()))
        os.Exit(1)
	}

    application := app.New(log, cfg, storage, cache, kafkaProducer, userClient, referralClient)

    // Используем WaitGroup для ожидания завершения всех горутин
    var wg sync.WaitGroup

    // Запуск gRPC сервера
    wg.Add(1)
    go func() {
        defer wg.Done()
        if err := application.GRPCServer.Run(); err != nil {
            log.Error("gRPC server error", slog.String("error", err.Error()))
        }
    }()

    // Запуск Kafka консьюмера
    wg.Add(1)
    go func() {
        defer wg.Done()
        kafkaConsumer.RunConsumer(ctx)
    }()

    // Ожидание сигнала для завершения
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

    <-stop
    log.Info("Received shutdown signal, initiating graceful shutdown")

    // Отмена контекста для сигнализации всем горутинам о необходимости завершения
    cancel()

    // Graceful shutdown для gRPC сервера
    application.GRPCServer.Stop()

    // Ожидание завершения всех горутин
    wg.Wait()

    log.Info("All goroutines have finished, closing resources")

    // Закрытие ресурсов (defer'ы сработают здесь)

    log.Info("Application stopped gracefully")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}