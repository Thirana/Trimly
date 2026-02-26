package store

import (
	"context"
	"errors"
	"time"
)

type Link struct {
	Code    string
	LongURL string
	// ExpiresAt is optional; nil means no expiry.
	ExpiresAt *time.Time
}

var ErrCodeExists = errors.New("code already exists")
var ErrIntentExists = errors.New("create intent already exists")

// LinkStore defines the behavior the shortener service needs.
// Any type that provides these methods (with exact signatures) satisfies it.
type LinkStore interface {
	Save(ctx context.Context, link Link) error
	Get(ctx context.Context, code string) (Link, bool, error)
	FindByIntent(ctx context.Context, longURL string, expiresAt *time.Time) (Link, bool, error)
}
