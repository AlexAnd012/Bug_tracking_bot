package filter_from_config

import (
	"Bug_tracking_bot/internal/log_processing"
	"fmt"
	"regexp"
	"strings"
)

type Matcher struct {
	allowedLevels map[string]struct{} // Если пустой, значит все 3 уровня
	alertRegex    []*regexp.Regexp
}

func NewMatcher(levels []string, patterns []string) (*Matcher, error) {
	m := &Matcher{
		allowedLevels: make(map[string]struct{}),
		alertRegex:    make([]*regexp.Regexp, 0, len(patterns)),
	}

	// Нормализация уровней
	for _, lvl := range levels {
		lvl = strings.ToUpper(strings.TrimSpace(lvl))
		if lvl != "" {
			m.allowedLevels[lvl] = struct{}{}
		}
	}

	// Компиляция регулярных выражений
	for _, p := range patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			return nil, fmt.Errorf("ошибка компиляции регулярного выражения %q: %w", p, err)
		}
		m.alertRegex = append(m.alertRegex, re)
	}

	return m, nil
}

func (m *Matcher) Match(entry log_processing.LogEntry) bool {
	// фильтр по уровню логов
	if len(m.allowedLevels) > 0 {
		if _, ok := m.allowedLevels[strings.ToUpper(entry.Level)]; !ok {
			return false
		}
	}
	// проверка регулярного выражения по сообщению
	if len(m.alertRegex) == 0 {
		return false
	}

	for _, re := range m.alertRegex {
		if re.MatchString(entry.Message) {
			return true
		}
	}
	return false
}
