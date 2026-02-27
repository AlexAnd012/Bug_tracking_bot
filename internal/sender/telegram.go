package sender

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"Bug_tracking_bot/internal/config"
)

type TelegramSender struct {
	token   string
	chatID  string
	client  *http.Client
	baseURL string
}

func NewTelegramSender(cfg config.TelegramConfig) (*TelegramSender, error) {
	if cfg.BotToken == "" || cfg.ChatID == "" {
		return nil, fmt.Errorf("отсутствует токен бота или ChatId")
	}

	return &TelegramSender{
		token:   cfg.BotToken,
		chatID:  cfg.ChatID,
		client:  &http.Client{Timeout: 5 * time.Second},
		baseURL: "https://api.telegram.org",
	}, nil
}

type telegramSendMessageRequest struct {
	ChatID                string `json:"chat_id"`
	Text                  string `json:"text"`
	ParseMode             string `json:"parse_mode,omitempty"` // для HTML
	DisableWebPagePreview bool   `json:"disable_web_page_preview,omitempty"`
}

type telegramSendMessageResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
}

func (s *TelegramSender) Send(ctx context.Context, text string) error {
	// Создаем тело запроса
	body := telegramSendMessageRequest{
		ChatID:                s.chatID,
		Text:                  text,
		ParseMode:             "HTML",
		DisableWebPagePreview: true,
	}

	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("ошибка кодирования в формат JSON: %w", err)
	}

	url := fmt.Sprintf("%s/bot%s/sendMessage", s.baseURL, s.token)

	// Создаем http запрос с контекстом
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("ошибка при создании запроса telegram: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("не удалось выполнить запрос в telegram: %w", err)
	}
	defer resp.Body.Close()

	// Читаем ответ telegram
	var tgResp telegramSendMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&tgResp); err != nil {
		return fmt.Errorf("декодируем ответ telegram: %w", err)
	}

	if resp.StatusCode != http.StatusOK || !tgResp.OK {
		return fmt.Errorf("ошибка api telegram: статус=%d ок=%v описание=%s", resp.StatusCode, tgResp.OK, tgResp.Description)
	}

	return nil
}
