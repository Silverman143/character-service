package kafkaconsumer

import (
	"context"
	"crypto/tls"
	"log/slog"
	"time"

	"github.com/Silverman143/character-service/internal/config"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/scram"
)

type KafkaConsumer struct {
    reader *kafka.Reader
    logger *slog.Logger
}


func getScramMechanism(username string, password string) (sasl.Mechanism, error) {
	return scram.Mechanism(scram.SHA512, username, password)
}

func createScramDialer(username string, password string) (*kafka.Dialer, error) {
	mechanism, err := getScramMechanism(username, password)
	if err != nil {
		return nil, err
	}
	return &kafka.Dialer{
		Timeout:       10 * time.Second,
		DualStack:     true,
		SASLMechanism: mechanism,
		TLS:           &tls.Config{},
	}, nil
}

func createScramTransport(username string, password string) (*kafka.Transport, error) {
	mechanism, err := getScramMechanism(username, password)
	if err != nil {
		return nil, err
	}
	return &kafka.Transport{
		DialTimeout: 10 * time.Second,
		SASL:        mechanism,
		TLS:         &tls.Config{},
	}, nil
}


func NewKafkaReader(cfg *config.KafkaConfig) (*kafka.Reader, error) {
	dialer, err := createScramDialer(cfg.User, cfg.Pass)
	if err != nil {
		return nil, err
	}

	conf := kafka.ReaderConfig{
		Brokers:     cfg.Broker,
		Topic:       cfg.TopicRead,
		GroupID:     cfg.GroupID,
		Dialer:      dialer,
	}

	return kafka.NewReader(conf), nil
}

func NewKafkaConsumer(cfg config.KafkaConfig,  log *slog.Logger) (*KafkaConsumer, error) {
    const op = "kafka.NewKafkaConsumer"

    reader, err := NewKafkaReader(&cfg)
    if err != nil{
        return nil, err
    }
    return &KafkaConsumer{
        reader: reader,
        logger: log,
    }, nil
}

func (c *KafkaConsumer) RunConsumer(ctx context.Context) {
	const op = "kafka.RunConsumer"
    logger := c.logger.With("op", op)

    logger.Info("Starting Kafka consumer")
    err := c.startConsumer(ctx)
    if err != nil {
        logger.Error("Kafka consumer stopped with error", slog.String("error", err.Error()))
    } else {
        logger.Info("Kafka consumer stopped")
    }
}

func (c *KafkaConsumer) startConsumer(ctx context.Context) error {
	const op = "kafka.startConsumer"
    logger := c.logger.With("op", op)
    logger.Info("Starting Kafka consumer")
    for {
        select {
        case <-ctx.Done():
            c.logger.Info("Kafka consumer stopped")
            return nil
        default:
            msg, err := c.reader.ReadMessage(ctx)
            if err != nil {
                logger.Error("Failed to read message", "error", err)
                continue
            }

            if err := c.HandleMessage(ctx, msg.Value); err != nil {
                logger.Error("Failed to handle message", "error", err)
                // Здесь можно добавить логику повторных попыток или обработки ошибок
            }
        }
    }
}

func (c *KafkaConsumer) Close() error {
    return c.reader.Close()
}
