package shortener

import "net/url"

// IsValidHTTPURL ensures:
// - parses correctly
// - has http/https scheme
// - has a host (so "http://example.com" is ok, "http:/x" is not)
func IsValidHTTPURL(raw string) bool {
	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	return u.Host != ""
}
