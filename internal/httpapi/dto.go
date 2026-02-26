package httpapi

import "time"

type CreateLinkRequest struct {
	LongURL   string     `json:"long_url" binding:"required"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type CreateLinkResponse struct {
	Code      string     `json:"code"`
	ShortURL  string     `json:"short_url"`
	LongURL   string     `json:"long_url"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type RedirectURI struct {
	Code string `uri:"code" binding:"required"`
}
