package parser

import (
	"Bug_tracking_bot/internal/log_processing"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var logLineRegexp = regexp.MustCompile(`^(\S+)\s+\[(DEBUG|INFO|ERROR)\]\s+(.+)$`)

func ParseLine(raw string) (log_processing.LogEntry, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return log_processing.LogEntry{}, fmt.Errorf("пустой лог")
	}

	// Возвращаем слайс, где m[0] — вся строка, m[1] — timestamp, m[2] — level, m[3] — message
	m := logLineRegexp.FindStringSubmatch(raw)
	if len(m) != 4 {
		return log_processing.LogEntry{}, fmt.Errorf("неверный формат лога")
	}

	var ts, err = time.Parse("2006-01-02T15:04:05Z07:00", m[1])
	if err != nil {
		return log_processing.LogEntry{}, fmt.Errorf("ошибка парсинга времени: %w", err)
	}

	return log_processing.LogEntry{
		Timestamp: ts,
		Level:     m[2],
		Message:   m[3],
		Raw:       raw,
	}, nil
}
