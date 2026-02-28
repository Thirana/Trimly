package store

import (
	"strings"
	"testing"
	"time"
)

func TestBuildIntentKey_NoNullByte(t *testing.T) {
	t.Parallel()

	expiresAt := time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)
	key := BuildIntentKey("https://example.com/path", &expiresAt)
	if strings.ContainsRune(key, '\x00') {
		t.Fatalf("intent key must not contain null byte: %q", key)
	}
}

func TestBuildIntentKey_StableForEquivalentExpiry(t *testing.T) {
	t.Parallel()

	utc := time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)
	local := utc.In(time.FixedZone("X", 5*60*60))

	k1 := BuildIntentKey("https://example.com/path", &utc)
	k2 := BuildIntentKey("https://example.com/path", &local)
	if k1 != k2 {
		t.Fatalf("expected stable key for equivalent times, got %q vs %q", k1, k2)
	}
}
