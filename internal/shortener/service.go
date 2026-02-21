package shortener

import (
	"context"
	"errors"

	"github.com/thirana/url-shortener/internal/store"
)

var ErrNotFound = errors.New("not found")
var ErrInvalidURL = errors.New("invalid url")

type Service struct {
	store store.LinkStore
}

func NewService(s store.LinkStore) *Service {
	return &Service{store: s}
}

func (s *Service) Create(ctx context.Context, longURL string) (store.Link, error) {
	if !IsValidHTTPURL(longURL) {
		return store.Link{}, ErrInvalidURL
	}

	// For now: generate a code once and store it.
	// Next steps we’ll add collision handling + retries.
	code, err := NewCode(7) // 7 chars is a good starting point
	if err != nil {
		return store.Link{}, err
	}

	link := store.Link{Code: code, LongURL: longURL}
	if err := s.store.Save(ctx, link); err != nil {
		return store.Link{}, err
	}
	return link, nil
}

func (s *Service) Resolve(ctx context.Context, code string) (store.Link, error) {
	link, ok, err := s.store.Get(ctx, code)
	if err != nil {
		return store.Link{}, err
	}
	if !ok {
		return store.Link{}, ErrNotFound
	}
	return link, nil
}
