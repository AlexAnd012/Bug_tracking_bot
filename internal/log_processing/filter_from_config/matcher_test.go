package filter_from_config

import (
	"Bug_tracking_bot/internal/log_processing"
	"testing"
	"time"
)

func TestMatcher_Match_ByRegexAndLevel(t *testing.T) {
	m, err := NewMatcher(
		[]string{"ERROR"},
		[]string{"^Error processing request$", "^Invalid input received$"},
	)
	if err != nil {
		t.Fatalf("ожидается без ошибок, получено: %v", err)
	}

	entry := log_processing.LogEntry{
		Timestamp: time.Now(),
		Level:     "ERROR",
		Message:   "Error processing request",
		Raw:       "2026-02-25T17:24:25+03:00 [ERROR] Error processing request",
	}

	if !m.Match(entry) {
		t.Fatal("Ожидается true, получено false")
	}
}

func TestMatcher_NoMatch_WrongRegex(t *testing.T) {
	m, err := NewMatcher(
		[]string{"ERROR"},
		[]string{"^Error processing request$"},
	)
	if err != nil {
		t.Fatalf("ожидалось без ошибок, получено: %v", err)
	}

	entry := log_processing.LogEntry{
		Timestamp: time.Now(),
		Level:     "ERROR",
		Message:   "User logged in",
		Raw:       "2026-02-25T17:24:25+03:00 [ERROR] User logged in",
	}

	if m.Match(entry) {
		t.Fatal("Ожидается true, получено false")
	}
}

func TestMatcher_EmptyLevels_AllLevelsAllowed(t *testing.T) {
	m, err := NewMatcher(
		[]string{},
		[]string{"^Invalid input received$"},
	)
	if err != nil {
		t.Fatalf("ожидается без ошибок, получено: %v", err)
	}

	entry := log_processing.LogEntry{
		Timestamp: time.Now(),
		Level:     "DEBUG",
		Message:   "Invalid input received",
		Raw:       "2026-02-25T17:24:25+03:00 [DEBUG] Invalid input received",
	}

	if !m.Match(entry) {
		t.Fatal("Ожидается true для пустого массива уровней")
	}
}

func TestMatcher_LevelFilteredOut(t *testing.T) {
	m, err := NewMatcher(
		[]string{"ERROR"},
		[]string{"^Invalid input received$"},
	)
	if err != nil {
		t.Fatalf("ожидается без ошибок, получено: %v", err)
	}

	entry := log_processing.LogEntry{
		Timestamp: time.Now(),
		Level:     "INFO",
		Message:   "Invalid input received",
		Raw:       "2026-02-25T17:24:25+03:00 [INFO] Invalid input received",
	}

	if m.Match(entry) {
		t.Fatal("ожидается false для некорректного уровня логов")
	}
}

func TestMatcher_InvalidRegex(t *testing.T) {
	_, err := NewMatcher(
		[]string{"ERROR"},
		[]string{"("}, // invalid regex
	)
	if err == nil {
		t.Fatal("ожидается ошибка для неверного регулярного выражения , получено nil")
	}
}
