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
}

func NewFileReader(path string) *FileReader {
	return &FileReader{path: path}
}

// ReadNewLines читаем только новые строки с прошлого вызова
func (r *FileReader) ReadNewLines() ([]string, error) {
	f, err := os.Open(r.path)
	if err != nil {
		return nil, fmt.Errorf("ошибка при открытии файла с логами: %w", err)
	}
	defer f.Close()

	// Получение размера файла
	info, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении данных о файле: %w", err)
	}

	// Если файл обрезали, начинаем сначала
	if info.Size() < r.alreadyRead {
		r.alreadyRead = 0
	}

	if _, err := f.Seek(r.alreadyRead, io.SeekStart); err != nil {
		return nil, fmt.Errorf("ошибка при переходе к позиции чтения: %w", err)
	}

	// Читаем строку и увеличиваем буфер сканера на случай длинных строк
	scanner := bufio.NewReader(f)

	var lines []string
	for {
		line, err := scanner.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				line = strings.TrimRight(line, "\r\n")
				if line != "" {
					lines = append(lines, line)
				}
				break
			}
			return nil, fmt.Errorf("ошибка чтения строк: %w", err)
		}
		line = strings.TrimRight(line, "\r\n")
		if line != "" {
			lines = append(lines, line)
		}
	}

	// после чтения сохраняем текущую позицию как новый alreadyRead
	pos, err := f.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения нового курсора чтения: %w", err)
	}
	r.alreadyRead = pos

	return lines, nil
}
