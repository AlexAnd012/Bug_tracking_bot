package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	LogFile        string         `yaml:"log_file"`
	PollIntervalMS int            `yaml:"poll_interval_ms"`
	Sender         Sender         `yaml:"sender"`
	Telegram       TelegramConfig `yaml:"telegram"`
	Filters        FiltersConfig  `yaml:"filters"`
	Format         FormatConfig   `yaml:"format"`
}

type Sender struct {
	Type string `yaml:"type"` // stdout | telegram
}

type TelegramConfig struct {
	BotToken string `yaml:"bot_token"`
	ChatID   string `yaml:"chat_id"`
}

type FiltersConfig struct {
	Levels     []string `yaml:"levels"`      // Если пустой, значит все 3 уровня
	AlertRegex []string `yaml:"alert_regex"` // Обязательные регулярные выражения для отбора логов
}

type FormatConfig struct {
	IncludeRaw         bool `yaml:"include_raw"`
	IncludeFingerprint bool `yaml:"include_fingerprint"`
}

func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ошибка при чтении конфигурации: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("ошибка декодирования yaml: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	if strings.TrimSpace(c.LogFile) == "" {
		return fmt.Errorf("отсутсвует файл с логами")
	}
	if c.PollIntervalMS <= 0 {
		c.PollIntervalMS = 500
	}

	c.Sender.Type = strings.ToLower(strings.TrimSpace(c.Sender.Type))
	switch c.Sender.Type {
	case "stdout", "telegram":
	default:
		return fmt.Errorf("config: тип отправителя должен быть stdout|telegram")
	}

	for i := range c.Filters.Levels {
		c.Filters.Levels[i] = strings.ToUpper(strings.TrimSpace(c.Filters.Levels[i]))
	}

	if c.Sender.Type == "telegram" {
		if c.Telegram.BotToken == "" || c.Telegram.ChatID == "" {
			return fmt.Errorf("config: отсутствует токен телеграм бота и ChatId")
		}
	}

	return nil
}
