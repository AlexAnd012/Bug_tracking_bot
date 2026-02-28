package formatter

import (
	"Bug_tracking_bot/internal/config"
	"Bug_tracking_bot/internal/log_processing"
	"Bug_tracking_bot/internal/log_processing/protect_from_duplicates"
	"fmt"
	"html"
)

func FormatTelegram(entry log_processing.LogEntry, cfg config.FormatConfig) string {
	var text string
	time := entry.Timestamp.Format("2006-01-02 15:04:05")
	level := html.EscapeString(entry.Level)
	msg := html.EscapeString(entry.Message)
	fp := html.EscapeString(protect_from_duplicates.Fingerprint(entry.Raw))
	raw := html.EscapeString(entry.Raw)

	switch level {
	case "INFO":
		text += "🟢"
	case "DEBUG":
		text += "🟡"
	case "ERROR":
		text += "🔴"
	}

	text += fmt.Sprintf(
		"<b> Уровень </b>%s\n\n"+
			"<b>Время:</b> %s\n\n"+
			"<b>Сообщение:</b> %s\n\n",
		level, time, msg,
	)

	if cfg.IncludeFingerprint {
		text += fmt.Sprintf("<b>Уникальный ключ:</b> <code>%s</code>\n\n", fp)
	}

	if cfg.IncludeRaw {
		text += fmt.Sprintf("<b>Исходный лог:</b>\n<code>%s</code>\n\n", raw)
	}

	return text
}
