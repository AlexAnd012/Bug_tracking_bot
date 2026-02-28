package main

import (
	"Bug_tracking_bot/internal/config"
	"Bug_tracking_bot/internal/log_processing/filter_from_config"
	"Bug_tracking_bot/internal/sender"
	"log"
	"time"
)

type ReloadResult struct {
	Applied             bool
	LogFileChanged      bool
	PollIntervalChanged bool
}

// Смотрим ModTime конфига, если изменился, то загружаем новый, собираем новый matcher и sender, подменяем их в rt
// Если поменялся путь к логу, то пересоздаём fileReader, чтобы читать новый файл.
func tryReloadRuntime(rt *Runtime) (ReloadResult, error) {
	mt, err := configModTime(rt.cfgPath)
	if err != nil {
		return ReloadResult{}, err
	}

	if !mt.After(rt.cfgMTime) {
		return ReloadResult{}, nil
	}

	log.Println("Обнаружено изменение config.yaml, проверяем...")

	newCfg, err := config.Load(rt.cfgPath)
	if err != nil {
		log.Printf("Ошибка загрузки YAML, конфиг не применён: %v", err)
		return ReloadResult{}, nil
	}

	newMatcher, err := filter_from_config.NewMatcher(newCfg.Filters.Levels, newCfg.Filters.AlertRegex)
	if err != nil {
		log.Printf("Ошибка компиляции regex, конфиг не применён: %v", err)
		return ReloadResult{}, nil
	}

	newSender, err := sender.New(newCfg)
	if err != nil {
		log.Printf("Ошибка создания sender, конфиг не применён: %v", err)
		return ReloadResult{}, nil
	}

	result := ReloadResult{
		Applied:             true,
		LogFileChanged:      rt.cfg.LogFile != newCfg.LogFile,
		PollIntervalChanged: rt.cfg.PollIntervalMS != newCfg.PollIntervalMS,
	}

	rt.cfg = newCfg
	rt.matcher = newMatcher
	rt.sender = newSender
	rt.cfgMTime = mt

	log.Println("Новый конфиг успешно применён")

	return result, nil
}

func newPollTicker(pollIntervalMS int) *time.Ticker {
	return time.NewTicker(time.Duration(pollIntervalMS) * time.Millisecond)
}
