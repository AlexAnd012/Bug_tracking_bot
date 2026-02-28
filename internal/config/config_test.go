package config

import "testing"

func TestConfig_Validate_ErrorWhenAlertRegexEmpty(t *testing.T) {
	cfg := &Config{
		LogFile:        "logs.log",
		PollIntervalMS: 1000,
		Sender: Sender{
			Type: "stdout",
		},
		Filters: FiltersConfig{
			Levels:     []string{"ERROR"},
			AlertRegex: []string{},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Ожидается ошибка, если filters.alert_regex пустой")
	}
}

func TestConfig_Validate_ErrorWhenAlertRegexMissing(t *testing.T) {
	cfg := &Config{
		LogFile:        "logs.log",
		PollIntervalMS: 1000,
		Sender: Sender{
			Type: "stdout",
		},
		Filters: FiltersConfig{
			Levels:     []string{"ERROR"},
			AlertRegex: nil,
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Ожидается ошибка, если filters.alert_regex не задан")
	}
}

func TestConfig_Validate_ErrorWhenAlertRegexContainsEmptyValue(t *testing.T) {
	cfg := &Config{
		LogFile:        "logs.log",
		PollIntervalMS: 1000,
		Sender: Sender{
			Type: "stdout",
		},
		Filters: FiltersConfig{
			Levels:     []string{"ERROR"},
			AlertRegex: []string{"ERROR", "   "},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Ожидается ошибка, если filters.alert_regex содержит пустое значение")
	}
}

func TestConfig_Validate_SuccessWhenAlertRegexFilled(t *testing.T) {
	cfg := &Config{
		LogFile:        "logs.log",
		PollIntervalMS: 1000,
		Sender: Sender{
			Type: "stdout",
		},
		Filters: FiltersConfig{
			Levels:     []string{"ERROR"},
			AlertRegex: []string{"Error", "Invalid input"},
		},
	}

	err := cfg.Validate()
	if err != nil {
		t.Fatalf("Не ожидалась ошибка для валидного filters.alert_regex, получено: %v", err)
	}
}
