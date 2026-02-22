package store

import "context"

type Link struct {
	Code    string
	LongURL string
}

// LinkStore defines the behavior the shortener service needs.
// Any type that provides these methods (with exact signatures) satisfies it.
type LinkStore interface {
	Save(ctx context.Context, link Link) error
	Get(ctx context.Context, code string) (Link, bool, error)
}
