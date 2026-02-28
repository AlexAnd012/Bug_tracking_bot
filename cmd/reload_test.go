package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeConfigFile(t *testing.T, path string, body string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("не удалось записать config.yaml: %v", err)
	}
}

func mustBuildRuntimeForTest(t *testing.T, path string) *Runtime {
	t.Helper()

	rt, err := buildRuntime(path)
	if err != nil {
		t.Fatalf("не удалось собрать runtime: %v", err)
	}

	// Если cfgPath не проставляется внутри buildRuntime, выставляем явно.
	rt.cfgPath = path

	return rt
}

func waitForMTimeTick() {
	time.Sleep(1100 * time.Millisecond)
}

func TestTryReloadRuntime_AppliesPollIntervalAndTelegramSettings(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	logPath1 := filepath.Join(dir, "logs-1.log")
	logPath2 := filepath.Join(dir, "logs-2.log")

	initialCfg := `
log_file: "` + logPath1 + `"
poll_interval_ms: 1000

sender:
  type: telegram

telegram:
  bot_token: "token-old"
  chat_id: "100"

filters:
  levels: ["ERROR"]
  alert_regex: ["Error", "Invalid"]

format:
  include_raw: true
  include_fingerprint: true
`

	writeConfigFile(t, cfgPath, initialCfg)

	rt := mustBuildRuntimeForTest(t, cfgPath)
	oldSender := rt.sender

	if rt.cfg.PollIntervalMS != 1000 {
		t.Fatalf("ожидался исходный poll_interval_ms=1000, получено %d", rt.cfg.PollIntervalMS)
	}
	if rt.cfg.Telegram.BotToken != "token-old" {
		t.Fatalf("ожидался исходный token=token-old, получено %q", rt.cfg.Telegram.BotToken)
	}
	if rt.cfg.Telegram.ChatID != "100" {
		t.Fatalf("ожидался исходный chat_id=100, получено %q", rt.cfg.Telegram.ChatID)
	}

	waitForMTimeTick()

	updatedCfg := `
log_file: "` + logPath2 + `"
poll_interval_ms: 250

sender:
  type: telegram

telegram:
  bot_token: "token-new"
  chat_id: "200"

filters:
  levels: ["ERROR"]
  alert_regex: ["Error", "Invalid"]

format:
  include_raw: false
  include_fingerprint: false
`

	writeConfigFile(t, cfgPath, updatedCfg)

	res, err := tryReloadRuntime(rt)
	if err != nil {
		t.Fatalf("не ожидалась ошибка reload: %v", err)
	}

	if !res.Applied {
		t.Fatal("ожидалось, что новый конфиг будет применён")
	}

	if !res.PollIntervalChanged {
		t.Fatal("ожидалось, что PollIntervalChanged=true")
	}

	if !res.LogFileChanged {
		t.Fatal("ожидалось, что LogFileChanged=true")
	}

	if rt.cfg.PollIntervalMS != 250 {
		t.Fatalf("ожидался обновлённый poll_interval_ms=250, получено %d", rt.cfg.PollIntervalMS)
	}

	if rt.cfg.LogFile != logPath2 {
		t.Fatalf("ожидался новый log_file=%q, получено %q", logPath2, rt.cfg.LogFile)
	}

	if rt.cfg.Telegram.BotToken != "token-new" {
		t.Fatalf("ожидался новый telegram.bot_token=token-new, получено %q", rt.cfg.Telegram.BotToken)
	}

	if rt.cfg.Telegram.ChatID != "200" {
		t.Fatalf("ожидался новый telegram.chat_id=200, получено %q", rt.cfg.Telegram.ChatID)
	}

	if rt.sender == oldSender {
		t.Fatal("ожидалось, что sender будет пересоздан при успешном hot reload")
	}
}

func TestTryReloadRuntime_NoChanges_ReturnsNotApplied(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	logPath := filepath.Join(dir, "logs.log")

	cfg := `
log_file: "` + logPath + `"
poll_interval_ms: 500

sender:
  type: telegram

telegram:
  bot_token: "token-1"
  chat_id: "123"

filters:
  levels: ["ERROR"]
  alert_regex: ["Error"]

format:
  include_raw: true
  include_fingerprint: true
`

	writeConfigFile(t, cfgPath, cfg)

	rt := mustBuildRuntimeForTest(t, cfgPath)
	oldCfgMTime := rt.cfgMTime
	oldSender := rt.sender

	res, err := tryReloadRuntime(rt)
	if err != nil {
		t.Fatalf("не ожидалась ошибка reload без изменений: %v", err)
	}

	if res.Applied {
		t.Fatal("не ожидалось применение конфига без изменения mtime")
	}

	if res.PollIntervalChanged {
		t.Fatal("не ожидалось PollIntervalChanged без изменения конфига")
	}

	if res.LogFileChanged {
		t.Fatal("не ожидалось LogFileChanged без изменения конфига")
	}

	if rt.cfgMTime != oldCfgMTime {
		t.Fatal("mtime runtime не должен меняться без изменений файла")
	}

	if rt.sender != oldSender {
		t.Fatal("sender не должен пересоздаваться без изменений файла")
	}
}

func TestTryReloadRuntime_InvalidConfig_DoesNotApply(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	logPath := filepath.Join(dir, "logs.log")

	initialCfg := `
log_file: "` + logPath + `"
poll_interval_ms: 700

sender:
  type: telegram

telegram:
  bot_token: "valid-token"
  chat_id: "777"

filters:
  levels: ["ERROR"]
  alert_regex: ["Error"]

format:
  include_raw: true
  include_fingerprint: true
`

	writeConfigFile(t, cfgPath, initialCfg)

	rt := mustBuildRuntimeForTest(t, cfgPath)
	oldSender := rt.sender
	oldPoll := rt.cfg.PollIntervalMS
	oldToken := rt.cfg.Telegram.BotToken
	oldChatID := rt.cfg.Telegram.ChatID
	oldLogFile := rt.cfg.LogFile

	waitForMTimeTick()

	// Невалидный reload: пустой token
	invalidCfg := `
log_file: "` + logPath + `"
poll_interval_ms: 250

sender:
  type: telegram

telegram:
  bot_token: ""
  chat_id: "999"

filters:
  levels: ["ERROR"]
  alert_regex: ["Error"]

format:
  include_raw: false
  include_fingerprint: false
`

	writeConfigFile(t, cfgPath, invalidCfg)

	res, err := tryReloadRuntime(rt)
	if err != nil {
		t.Fatalf("не ожидалась фатальная ошибка tryReloadRuntime, даже если конфиг невалидный: %v", err)
	}

	if res.Applied {
		t.Fatal("не ожидалось применение невалидного конфига")
	}

	if rt.cfg.PollIntervalMS != oldPoll {
		t.Fatalf("poll_interval_ms не должен был измениться, ожидалось %d, получено %d", oldPoll, rt.cfg.PollIntervalMS)
	}

	if rt.cfg.Telegram.BotToken != oldToken {
		t.Fatalf("telegram.bot_token не должен был измениться, ожидалось %q, получено %q", oldToken, rt.cfg.Telegram.BotToken)
	}

	if rt.cfg.Telegram.ChatID != oldChatID {
		t.Fatalf("telegram.chat_id не должен был измениться, ожидалось %q, получено %q", oldChatID, rt.cfg.Telegram.ChatID)
	}

	if rt.cfg.LogFile != oldLogFile {
		t.Fatalf("log_file не должен был измениться, ожидалось %q, получено %q", oldLogFile, rt.cfg.LogFile)
	}

	if rt.sender != oldSender {
		t.Fatal("sender не должен пересоздаваться, если reload не применился")
	}
}

func TestTryReloadRuntime_PollIntervalOnlyChange_SetsOnlyPollFlag(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	logPath := filepath.Join(dir, "logs.log")

	initialCfg := `
log_file: "` + logPath + `"
poll_interval_ms: 500

sender:
  type: telegram

telegram:
  bot_token: "token-a"
  chat_id: "1"

filters:
  levels: ["ERROR"]
  alert_regex: ["Error"]

format:
  include_raw: true
  include_fingerprint: true
`

	writeConfigFile(t, cfgPath, initialCfg)

	rt := mustBuildRuntimeForTest(t, cfgPath)

	waitForMTimeTick()

	updatedCfg := `
log_file: "` + logPath + `"
poll_interval_ms: 900

sender:
  type: telegram

telegram:
  bot_token: "token-a"
  chat_id: "1"

filters:
  levels: ["ERROR"]
  alert_regex: ["Error"]

format:
  include_raw: true
  include_fingerprint: true
`

	writeConfigFile(t, cfgPath, updatedCfg)

	res, err := tryReloadRuntime(rt)
	if err != nil {
		t.Fatalf("не ожидалась ошибка reload: %v", err)
	}

	if !res.Applied {
		t.Fatal("ожидалось применение обновлённого конфига")
	}

	if !res.PollIntervalChanged {
		t.Fatal("ожидалось PollIntervalChanged=true")
	}

	if res.LogFileChanged {
		t.Fatal("не ожидалось LogFileChanged при неизменном log_file")
	}

	if rt.cfg.PollIntervalMS != 900 {
		t.Fatalf("ожидался poll_interval_ms=900, получено %d", rt.cfg.PollIntervalMS)
	}
}
