package reader

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileReader_ReadNewLines_WaitsForPartialLineCompletion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "logs.log")

	// Сначала пишем только неполную строку без \n
	if err := os.WriteFile(path, []byte("2026-02-28T12:00:00Z [ERROR] Datab"), 0o644); err != nil {
		t.Fatalf("не удалось создать тестовый лог: %v", err)
	}

	r := NewFileReader(path)

	// Неполная строка не должна считаться готовым логом
	lines, err := r.ReadNewLines()
	if err != nil {
		t.Fatalf("не ожидалась ошибка при первом чтении: %v", err)
	}
	if len(lines) != 0 {
		t.Fatalf("ожидалось 0 строк для неполной записи, получено: %v", lines)
	}

	// Дозаписываем хвост и перевод строки
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("не удалось открыть файл для дозаписи: %v", err)
	}

	if _, err := f.WriteString("ase connection failed\n"); err != nil {
		_ = f.Close()
		t.Fatalf("не удалось дозаписать хвост строки: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("не удалось закрыть файл после дозаписи: %v", err)
	}

	// Теперь должна прийти целая, склеенная строка
	lines, err = r.ReadNewLines()
	if err != nil {
		t.Fatalf("не ожидалась ошибка при втором чтении: %v", err)
	}

	if len(lines) != 1 {
		t.Fatalf("ожидалась 1 завершённая строка, получено %d: %v", len(lines), lines)
	}

	want := "2026-02-28T12:00:00Z [ERROR] Database connection failed"
	if lines[0] != want {
		t.Fatalf("ожидалась строка %q, получено %q", want, lines[0])
	}

	// Повторно та же строка приходить не должна
	lines, err = r.ReadNewLines()
	if err != nil {
		t.Fatalf("не ожидалась ошибка при третьем чтении: %v", err)
	}
	if len(lines) != 0 {
		t.Fatalf("ожидалось 0 строк после повторного чтения, получено: %v", lines)
	}
}
