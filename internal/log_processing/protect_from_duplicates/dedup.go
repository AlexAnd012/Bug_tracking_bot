package protect_from_duplicates

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

type Deduplicator struct {
	seen     map[string]time.Time
	timeLife time.Duration // Время жизни лога в памяти, если встречается раньше чем через timeLife, то блокируем лог
}

func NewDeduplicator(ttl time.Duration) *Deduplicator {
	return &Deduplicator{
		seen:     make(map[string]time.Time),
		timeLife: ttl,
	}
}

// Allow возвращает true, если лог еще не отправлялся недавно.
func (d *Deduplicator) Allow(raw string) bool {
	now := time.Now()
	key := Fingerprint(raw)

	// периодическая очистка через timeLife
	for k, t := range d.seen {
		if now.Sub(t) > d.timeLife {
			delete(d.seen, k)
		}
	}

	if t, ok := d.seen[key]; ok && now.Sub(t) <= d.timeLife {
		return false
	}

	d.seen[key] = now
	return true
}

func Fingerprint(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])[:12] // короткий хэш для хранения в мапе
}
