package parser

import (
	"testing"
	"time"
)

func TestParseLine_Valid(t *testing.T) {
	raw := "2026-02-25T17:24:25+03:00 [DEBUG] Invalid input received"

	entry, err := ParseLine(raw)
	if err != nil {
		t.Fatalf("Ожидается без ошибок, получена ошибка: %v", err)
	}

	if entry.Level != "DEBUG" {
		t.Fatalf("ожидается уровень DEBUG, получен: %s", entry.Level)
	}

	if entry.Message != "Invalid input received" {
		t.Fatalf("ожидается сообщение %q, получено: %q", "Invalid input received", entry.Message)
	}

	if entry.Raw != raw {
		t.Fatalf("ожидалась исходная строка %q, получено: %q", raw, entry.Raw)
	}

	wantTime, _ := time.Parse(time.RFC3339, "2026-02-25T17:24:25+03:00")
	if !entry.Timestamp.Equal(wantTime) {
		t.Fatalf("ожидалось время %v, получено %v", wantTime, entry.Timestamp)
	}
}

func TestParseLine_EmptyLine(t *testing.T) {
	_, err := ParseLine("   ")
	if err == nil {
		t.Fatal("ожидается ошибка на пустую строку, получено nil")
	}
}

func TestParseLine_InvalidFormat(t *testing.T) {
	raw := "невалидный лог"

	_, err := ParseLine(raw)
	if err == nil {
		t.Fatal("ожидалась ошибка из-за невалидного формата, получено nil")
	}
}

func TestParseLine_InvalidTimestamp(t *testing.T) {
	raw := "невалидное_время [ERROR] Error processing request"

	_, err := ParseLine(raw)
	if err == nil {
		t.Fatal("ожидается ошибка невалидного времени, получено nil")
	}
}
