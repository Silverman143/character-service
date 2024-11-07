package kafkaproducer

import (
	"crypto/tls"
	"log/slog"
	"time"

	"github.com/Silverman143/character-service/internal/config"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
)

type KafkaProducer struct {
    writer *kafka.Writer
    logger *slog.Logger
}

func NewKafkaProducer(cfg config.KafkaConfig, log *slog.Logger) (*KafkaProducer, error) {
	const op = "kafka.NewKafkaProducer"
	logger := log.With("op", op)

    mechanism, err := scram.Mechanism(scram.SHA512, cfg.User, cfg.Pass)
    if err != nil{
        logger.Error("error with creating SASL ", "error", err)
        return nil, err
    }
    
    dialer := &kafka.Dialer{
        Timeout:       10 * time.Second,
        DualStack:     true,
        SASLMechanism: mechanism,
        TLS:          &tls.Config{
            MinVersion: tls.VersionTLS12,
        },
    }

    writer := kafka.NewWriter(kafka.WriterConfig{
        Brokers: cfg.Broker,
        Topic:   cfg.TopicWrite,
        Dialer:  dialer,
    })

    return &KafkaProducer{
        writer: writer,
        logger: log,
    }, nil
}

func (p *KafkaProducer) Close() error {
    return p.writer.Close()
}