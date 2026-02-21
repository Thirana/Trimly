package shortener

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
)

// NewCode creates a URL-safe code.
// We use crypto/rand for unpredictability (better if you ever add "private links").
// Later, for extreme throughput, we can swap to a faster deterministic ID strategy.
func NewCode(n int) (string, error) {
	// Generate slightly more bytes, then base64-url encode and trim.
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	s := base64.RawURLEncoding.EncodeToString(b) // URL-safe
	s = strings.TrimRight(s, "=")
	if len(s) > n {
		s = s[:n]
	}
	return s, nil
}
