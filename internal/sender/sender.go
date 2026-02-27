package sender

import (
	"Bug_tracking_bot/internal/config"
	"context"
	"fmt"
	"strings"
)

type Sender interface {
	Send(ctx context.Context, text string) error
}

func New(cfg *config.Config) (Sender, error) {
	switch strings.ToLower(cfg.Sender.Type) {
	case "stdout":
		return &StdoutSender{}, nil
	case "telegram":
		return NewTelegramSender(cfg.Telegram)
	default:
		return nil, fmt.Errorf("не поддерживаемый тип отправления данных: %s", cfg.Sender.Type)
	}
}
