package main

import (
	logproc "Bug_tracking_bot/internal/log_processing/formatter"
	"Bug_tracking_bot/internal/log_processing/parser"
	"Bug_tracking_bot/internal/log_processing/protect_from_duplicates"
	"Bug_tracking_bot/internal/reader"
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type LogReader interface {
	ReadNewLines() ([]string, error)
}

const configPath = "config.yaml"

func main() {
	rt, err := buildRuntime(configPath)
	if err != nil {
		log.Fatalf("Ошибка запуска: %v", err)
	}

	var fileReader LogReader = reader.NewFileReader(rt.cfg.LogFile)
	deDupl := protect_from_duplicates.NewDeduplicator(5 * time.Minute)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go watchShutdown(cancel)

	ticker := newPollTicker(rt.cfg.PollIntervalMS)
	defer ticker.Stop()

	reloadTicker := time.NewTicker(1 * time.Second)
	defer reloadTicker.Stop()

	log.Println("Старт работы Bug_tracking_bot")

	for {
		select {
		case <-ctx.Done():
			log.Println("Завершение работы Bug_tracking_bot")
			return

		case <-reloadTicker.C:
			ticker, fileReader = handleReload(rt, ticker, fileReader)

		case <-ticker.C:
			processBatch(ctx, rt, fileReader, deDupl)
		}
	}
}

func watchShutdown(cancel context.CancelFunc) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	log.Println("Получен сигнал остановки")
	cancel()
}

func handleReload(rt *Runtime, ticker *time.Ticker, fileReader LogReader) (*time.Ticker, LogReader) {
	res, err := tryReloadRuntime(rt)
	if err != nil {
		log.Printf("Ошибка перезагрузки конфига: %v", err)
		return ticker, fileReader
	}

	if !res.Applied {
		return ticker, fileReader
	}

	if res.LogFileChanged {
		fileReader = reader.NewFileReader(rt.cfg.LogFile)
	}

	if res.PollIntervalChanged {
		ticker.Stop()
		ticker = newPollTicker(rt.cfg.PollIntervalMS)
	}

	log.Printf(
		"config перезагружен: файл с логами = %s отправитель = %s Время между чтением логов = %dms",
		rt.cfg.LogFile,
		rt.cfg.Sender.Type,
		rt.cfg.PollIntervalMS,
	)

	return ticker, fileReader
}

func processBatch(ctx context.Context, rt *Runtime, fileReader LogReader, deDupl *protect_from_duplicates.Deduplicator,
) {
	lines, err := fileReader.ReadNewLines()
	if err != nil {
		log.Printf("ошибка чтения строк: %v", err)
		return
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

		var msg string
		if rt.cfg.Sender.Type == "stdout" {
			msg = logproc.FormatStdout(entry, rt.cfg.Format)
		}
		if rt.cfg.Sender.Type == "telegram" {
			msg = logproc.FormatTelegram(entry, rt.cfg.Format)
		}

		sendCtx, cancelSend := context.WithTimeout(ctx, 10*time.Second)
		err = rt.sender.Send(sendCtx, msg)
		cancelSend()

		if err == nil {
			continue
		}

		if errors.Is(err, context.Canceled) {
			continue
		}

		if errors.Is(err, context.DeadlineExceeded) {
			log.Printf("Ошибка отправки, не успели отправить за отведенное время: %v", err)
			continue
		}

		log.Printf("Ошибка отправки: %v", err)
	}
}
