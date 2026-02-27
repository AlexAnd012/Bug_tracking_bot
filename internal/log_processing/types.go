package log_processing

import "time"

// LogEntry Структура записи лога после парсинга
type LogEntry struct {
	Timestamp time.Time
	Level     string
	Message   string
	Raw       string
}
