package kafkaconsumer

import (
	"context"
	"encoding/json"
)

func (h *KafkaConsumer) HandleMessage(ctx context.Context, message []byte) error {
	const op = "kafka.controllers.HandleMessage"
	logger := h.logger.With("op", op);

    var event struct {
        Type string `json:"type"`
    }
    if err := json.Unmarshal(message, &event); err != nil {
        logger.Error("Failed to unmarshal event", "error", err)
        return err
    }

    switch event.Type {
    case "user_update":
        return h.HandleUserUpdateData(ctx, message)
    // Добавьте другие типы событий по мере необходимости
    default:
        logger.Warn("Unknown event type", "type", event.Type)
        return nil
    }
}

func (h *KafkaConsumer) HandleUserUpdateData(ctx context.Context, message []byte) error {
	const op = "kafka.controllers.HandleMessage"
	logger := h.logger.With("op", op);

    var event struct {
        UserID string `json:"user_id"`
        Data   interface{} `json:"data"`
    }
    if err := json.Unmarshal(message, &event); err != nil {
        logger.Error("Failed to unmarshal user update event", "error", err)
        return err
    }

    // err := h.userService.UpdateData(ctx, event.UserID, event.Data)
    // if err != nil {
    //     h.logger.Error("Failed to update user data", "error", err, "userID", event.UserID)
    //     return err
    // }

    logger.Info("Successfully updated user data", "userID", event.UserID)
    return nil
}