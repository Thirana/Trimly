package rediscache

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/thirana/url-shortener/internal/store"
)

func TestKeyFormat(t *testing.T) {
	t.Parallel()

	if got, want := shortKey("abc123"), "v1:url:short:abc123"; got != want {
		t.Fatalf("short key = %q, want %q", got, want)
	}
	if got, want := missKey("abc123"), "v1:url:miss:abc123"; got != want {
		t.Fatalf("miss key = %q, want %q", got, want)
	}
}

func TestEncodeShortValue(t *testing.T) {
	t.Parallel()

	expiresAt := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	raw, err := encodeShortValue(store.Link{
		Code:      "abc123",
		LongURL:   "https://example.com/path",
		ExpiresAt: &expiresAt,
	})
	if err != nil {
		t.Fatalf("encodeShortValue returned error: %v", err)
	}

	var payload cachedShortValue
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		t.Fatalf("unmarshal payload returned error: %v", err)
	}
	if payload.LongURL != "https://example.com/path" {
		t.Fatalf("payload long_url = %q, want %q", payload.LongURL, "https://example.com/path")
	}
	if payload.ExpiresAtNano == nil {
		t.Fatalf("payload expiry is nil, want value")
	}
}
