package main

import (
	"Bug_tracking_bot/internal/log_processing/filter_from_config"
	"Bug_tracking_bot/internal/log_processing/parser"
	"Bug_tracking_bot/internal/log_processing/protect_from_duplicates"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"Bug_tracking_bot/internal/config"
	logproc "Bug_tracking_bot/internal/log_processing"
	"Bug_tracking_bot/internal/reader"
	"Bug_tracking_bot/internal/sender"
)

type Runtime struct {
	cfg      *config.Config              // Конфиг
	matcher  *filter_from_config.Matcher // Скомпилированные фильтры
	sender   sender.Sender               // Отправитель
	cfgMTime time.Time                   // Время изменения конфига
	cfgPath  string                      // Путь к конфигу
}

func main() {
	const configPath = "config.yaml"

	rt, err := buildRuntime(configPath)
	if err != nil {
		log.Fatalf("startup error: %v", err)
	}
	rt.cfgPath = configPath

	// Создаем объект структуры FileReader, чтобы читать файл с начала
	fileReader := reader.NewFileReader(rt.cfg.LogFile)
	// Создаем объект структуры Deduplicator, и указываем, что дубли не будут отправляться в течении 5 минут
	deDupl := protect_from_duplicates.NewDeduplicator(5 * time.Minute)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Отдельная горутина ловит SIGINT/SIGTERM и отменяет ctx
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch
		log.Println("Получен сигнал остановки")
		cancel()
	}()

	// Тикер, который указывает как часто читать лог
	ticker := time.NewTicker(time.Duration(rt.cfg.PollIntervalMS) * time.Millisecond)
	defer ticker.Stop()

	// Тикер для hot reload конфига
	reloadTicker := time.NewTicker(1 * time.Second)
	defer reloadTicker.Stop()

	log.Println("Старт работы Bug_tracking_bot")

	for {
		select {
		case <-ctx.Done():
			log.Println("Завершение работы Bug_tracking_bot")
			return

		case <-reloadTicker.C:
			changed, err := tryReloadRuntime(rt)
			if err != nil {
				log.Printf("Ошибка перезагрузки конфига: %v", err)
			} else if changed {
				// если поменялся путь к логу, пересоздаем reader
				fileReader = reader.NewFileReader(rt.cfg.LogFile)
				log.Printf("config перезагружен: файл с логами = %s отправитель = %s Время между чтением логов = %dms",
					rt.cfg.LogFile, rt.cfg.Sender.Type, rt.cfg.PollIntervalMS)
			}

		case <-ticker.C:
			lines, err := fileReader.ReadNewLines()
			if err != nil {
				log.Printf("ошибка чтения строк: %v", err)
				continue
			}

			for _, line := range lines {
				entry, err := parser.ParseLine(line)
				if err != nil {
					continue
				}

				if !rt.matcher.Match(entry) {
					continue
				}

				if !deDupl.Allow(entry.Raw) {
					continue
				}

				msg := logproc.Format(entry, rt.cfg.Format)

				sendCtx, cancelSend := context.WithTimeout(ctx, 10*time.Second)
				err = rt.sender.Send(sendCtx, msg)
				cancelSend()

				if err != nil {
					// Если приложение завершается (Ctrl+C)
					if errors.Is(err, context.Canceled) {
						continue
					}
					// Если не успели отправить за timeout
					if errors.Is(err, context.DeadlineExceeded) {
						log.Printf("Ошибка отправки, не успели отправить за отведенное время: %v", err)
						continue
					}

					log.Printf("Ошибка отправки: %v", err)
					continue
				}
			}
		}
	}
}

// Загружаем конфиг, создаём matcher, создаём sender, запоминаем ModTime конфига, возвращаем объект структуры Runtime
func buildRuntime(configPath string) (*Runtime, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки config.yaml: %w", err)
	}

	// Валидация конфига (если добавлен Validate)
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("ошибка валидации config.yaml: %w", err)
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

	log.Printf("Config загружен: файл с логами = %s отправитель = %s Время между чтением логов = %dms", cfg.LogFile, cfg.Sender.Type, cfg.PollIntervalMS)
	log.Printf("Фильтр по уровням =%v (Если пусто, все уровни)", cfg.Filters.Levels)

	return &Runtime{
		cfg:      cfg,
		matcher:  matcher,
		sender:   snd,
		cfgMTime: mt,
	}, nil
}

// Смотрим ModTime конфига, если изменился, то загружаем новый, собираем новый matcher и sender, подменяем их в rt
// Если поменялся путь к логу, то пересоздаём fileReader, чтобы читать новый файл.
func tryReloadRuntime(rt *Runtime) (bool, error) {
	mt, err := configModTime(rt.cfgPath)
	if err != nil {
		return false, err
	}

	if !mt.After(rt.cfgMTime) {
		return false, nil
	}

	log.Println("Обнаружено изменение config.yaml, проверяем...")

	newCfg, err := config.Load(rt.cfgPath)
	if err != nil {
		log.Printf("Ошибка загрузки YAML, конфиг не применён: %v", err)
		return false, nil // не применяем
	}

	// Валидируем структуру
	if err := newCfg.Validate(); err != nil {
		log.Printf("Ошибка валидации конфига, конфиг не применён: %v", err)
		return false, nil
	}

	// Пробуем создать matcher
	newMatcher, err := filter_from_config.NewMatcher(newCfg.Filters.Levels, newCfg.Filters.AlertRegex)
	if err != nil {
		log.Printf("Ошибка компиляции regex, конфиг не применён: %v", err)
		return false, nil
	}

	// Пробуем создать sender
	newSender, err := sender.New(newCfg)
	if err != nil {
		log.Printf("Ошибка создания sender, конфиг не применён: %v", err)
		return false, nil
	}

	// если все успешно, то применяем
	rt.cfg = newCfg
	rt.matcher = newMatcher
	rt.sender = newSender
	rt.cfgMTime = mt

	log.Println("Новый конфиг успешно применён")

	return true, nil
}

// Функция для нахождения времени изменения конфига
func configModTime(path string) (time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}
