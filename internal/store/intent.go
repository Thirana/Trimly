package store

import (
	"strconv"
	"strings"
	"time"
)

// BuildIntentKey creates the canonical create-intent key used for idempotency.
// Key structure must stay stable across store implementations.
func BuildIntentKey(longURL string, expiresAt *time.Time) string {
	var b strings.Builder
	b.WriteString(strconv.Itoa(len(longURL)))
	b.WriteString(":")
	b.WriteString(longURL)
	b.WriteString("|")

	if expiresAt == nil {
		b.WriteString("no-expiry")
		return b.String()
	}
	b.WriteString(expiresAt.UTC().Format(time.RFC3339Nano))
	return b.String()
}
