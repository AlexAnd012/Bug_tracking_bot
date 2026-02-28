package main

import (
	"Bug_tracking_bot/internal/config"
	"Bug_tracking_bot/internal/log_processing/filter_from_config"
	"Bug_tracking_bot/internal/sender"
	"fmt"
	"log"
	"os"
	"time"
)

type Runtime struct {
	cfg      *config.Config
	matcher  *filter_from_config.Matcher
	sender   sender.Sender
	cfgMTime time.Time
	cfgPath  string
}

// Загружаем конфиг, создаём matcher, создаём sender, запоминаем ModTime конфига, возвращаем объект структуры Runtime
func buildRuntime(configPath string) (*Runtime, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки config.yaml: %w", err)
	}

	matcher, err := filter_from_config.NewMatcher(cfg.Filters.Levels, cfg.Filters.AlertRegex)
	if err != nil {
		return nil, fmt.Errorf("ошибка компиляции regex в config.yaml: %w", err)
	}

	snd, err := sender.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("ошибка инициализации sender (%s): %w", cfg.Sender.Type, err)
	}

	mt, err := configModTime(configPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения времени изменения config.yaml: %w", err)
	}

	log.Printf(
		"Config загружен: файл с логами = %s отправитель = %s Время между чтением логов = %dms",
		cfg.LogFile,
		cfg.Sender.Type,
		cfg.PollIntervalMS,
	)
	log.Printf("Фильтр по уровням = %v (если пусто, все уровни)", cfg.Filters.Levels)

	return &Runtime{
		cfg:      cfg,
		matcher:  matcher,
		sender:   snd,
		cfgMTime: mt,
		cfgPath:  configPath,
	}, nil
}

func configModTime(path string) (time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}
