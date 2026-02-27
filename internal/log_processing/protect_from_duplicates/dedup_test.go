package protect_from_duplicates

import (
	"testing"
	"time"
)

func TestDeduplicator_Allow_FirstTime(t *testing.T) {
	d := NewDeduplicator(1 * time.Second)

	if !d.Allow("same raw log") {
		t.Fatal("Ожидается true для первого лога")
	}
}

func TestDeduplicator_Block_DuplicateWithinTTL(t *testing.T) {
	d := NewDeduplicator(1 * time.Second)

	raw := "same raw log"

	if !d.Allow(raw) {
		t.Fatal("Ожидается true для первого лога")
	}

	if d.Allow(raw) {
		t.Fatal("Ожидается false для повторного лога")
	}
}

func TestDeduplicator_Allow_AfterTTL(t *testing.T) {
	d := NewDeduplicator(50 * time.Millisecond)

	raw := "same raw log"

	if !d.Allow(raw) {
		t.Fatal("Ожидается true для первого лога")
	}

	time.Sleep(60 * time.Millisecond)

	if !d.Allow(raw) {
		t.Fatal("Ожидается true для повторного лога через время жизни лога")
	}
}

func TestFingerprint_SameInput_SameOutput(t *testing.T) {
	raw := "2026-02-25T17:24:25+03:00 [DEBUG] Invalid input received"

	fp1 := Fingerprint(raw)
	fp2 := Fingerprint(raw)

	if fp1 != fp2 {
		t.Fatalf("Ожидается одинаковый уникальный ключ, получено %q и %q", fp1, fp2)
	}

	if len(fp1) != 12 {
		t.Fatalf("ожидается уникальный ключ длины 12, получено %d", len(fp1))
	}
}
