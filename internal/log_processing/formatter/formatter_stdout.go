package formatter

import (
	"Bug_tracking_bot/internal/config"
	"Bug_tracking_bot/internal/log_processing"
	"Bug_tracking_bot/internal/log_processing/protect_from_duplicates"
	"fmt"
)

func FormatStdout(entry log_processing.LogEntry, cfg config.FormatConfig) string {
	var text string

	time := entry.Timestamp.Format("2006-01-02 15:04:05")
	level := entry.Level
	msg := entry.Message
	fp := protect_from_duplicates.Fingerprint(entry.Raw)
	raw := entry.Raw

	text += fmt.Sprintf(
		"Уровень: %s\n"+
			"Время: %s\n"+
			"Сообщение: %s\n",
		level, time, msg,
	)

	if cfg.IncludeFingerprint {
		text += fmt.Sprintf("Уникальный ключ: %s\n", fp)
	}

	if cfg.IncludeRaw {
		text += fmt.Sprintf("Исходный лог: %s\n", raw)
	}

	return text
}
