package store

import "context"

type Link struct {
	Code    string
	LongURL string
}

type LinkStore interface {
	Save(ctx context.Context, link Link) error
	Get(ctx context.Context, code string) (Link, bool, error)
}
