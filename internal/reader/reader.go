package reader

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type FileReader struct {
	path        string
	alreadyRead int64
	pendingTail string
}

func NewFileReader(path string) *FileReader {
	return &FileReader{path: path}
}

// ReadNewLines читает только завершённые новые строки с прошлого вызова.
// Неполная последняя строка без '\n' сохраняется и дочитывается на следующем вызове.
func (r *FileReader) ReadNewLines() ([]string, error) {
	f, err := os.Open(r.path)
	if err != nil {
		return nil, fmt.Errorf("ошибка при открытии файла с логами: %w", err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении данных о файле: %w", err)
	}

	// Если файл обрезали, начинаем сначала и сбрасываем хвост.
	if info.Size() < r.alreadyRead {
		r.alreadyRead = 0
		r.pendingTail = ""
	}

	if _, err := f.Seek(r.alreadyRead, io.SeekStart); err != nil {
		return nil, fmt.Errorf("ошибка при переходе к позиции чтения: %w", err)
	}

	reader := bufio.NewReader(f)
	var lines []string

	for {
		line, err := reader.ReadString('\n')

		switch {
		case err == nil:
			// Полная строка. Если был сохранён хвост — склеиваем.
			if r.pendingTail != "" {
				line = r.pendingTail + line
				r.pendingTail = ""
			}

			line = strings.TrimRight(line, "\r\n")
			if line != "" {
				lines = append(lines, line)
			}

		case err == io.EOF:
			// Неполную строку не отдаём наружу — сохраняем до следующего чтения.
			if line != "" {
				r.pendingTail += line
			}
			// В alreadyRead сохраняем позицию ДО pendingTail,
			// чтобы в следующий раз перечитать хвост и корректно склеить.
			pos, seekErr := f.Seek(0, io.SeekCurrent)
			if seekErr != nil {
				return nil, fmt.Errorf("ошибка получения нового курсора чтения: %w", seekErr)
			}
			r.alreadyRead = pos - int64(len(line))
			return lines, nil

		default:
			return nil, fmt.Errorf("ошибка чтения строк: %w", err)
		}
	}
}
