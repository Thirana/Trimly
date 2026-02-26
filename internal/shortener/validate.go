package shortener

import (
	"errors"
	"net"
	"net/url"
	"strings"
)

// IsValidHTTPURL ensures:
// - parses correctly
// - has http/https scheme
// - has a host (so "http://example.com" is ok, "http:/x" is not)
func IsValidHTTPURL(raw string) bool {
	_, err := NormalizeHTTPURL(raw)
	return err == nil
}

// NormalizeHTTPURL validates and canonicalizes a URL for idempotency checks.
// Current normalization:
// - trim spaces
// - enforce http/https + host
// - lowercase scheme and host
// - remove default ports (80/443)
// - normalize empty path to "/"
func NormalizeHTTPURL(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	u, err := url.ParseRequestURI(trimmed)
	if err != nil {
		return "", err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", errors.New("scheme must be http or https")
	}
	if u.Host == "" {
		return "", errors.New("host is required")
	}

	u.Scheme = strings.ToLower(u.Scheme)
	host := strings.ToLower(u.Hostname())
	port := u.Port()
	switch {
	case u.Scheme == "http" && port == "80":
		port = ""
	case u.Scheme == "https" && port == "443":
		port = ""
	}
	if port == "" {
		u.Host = host
	} else {
		u.Host = net.JoinHostPort(host, port)
	}
	if u.Path == "" {
		u.Path = "/"
	}
	return u.String(), nil
}
