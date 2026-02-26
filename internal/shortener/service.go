package shortener

import (
	"context"
	"errors"
	"time"

	"github.com/thirana/url-shortener/internal/store"
)

var ErrNotFound = errors.New("not found")
var ErrInvalidURL = errors.New("invalid url")
var ErrExpired = errors.New("expired")
var ErrCollision = errors.New("collision")

const (
	defaultCodeLength  = 7
	defaultMaxAttempts = 5
)

type Service struct {
	store       store.LinkStore
	codeLength  int
	maxAttempts int
	newCode     func(int) (string, error)
	now         func() time.Time
}

func NewService(s store.LinkStore) *Service {
	return &Service{
		store:       s,
		codeLength:  defaultCodeLength,
		maxAttempts: defaultMaxAttempts,
		newCode:     NewCode,
		now:         time.Now,
	}
}

func (s *Service) Create(ctx context.Context, longURL string, expiresAt *time.Time) (store.Link, bool, error) {
	normalizedURL, err := NormalizeHTTPURL(longURL)
	if err != nil {
		return store.Link{}, false, ErrInvalidURL
	}

	normalizedExpiry, err := s.normalizeExpiry(expiresAt)
	if err != nil {
		return store.Link{}, false, err
	}

	existing, ok, err := s.store.FindByIntent(ctx, normalizedURL, normalizedExpiry)
	if err != nil {
		return store.Link{}, false, err
	}
	if ok {
		return existing, false, nil
	}

	for attempts := 0; attempts < s.maxAttempts; attempts++ {
		code, codeErr := s.newCode(s.codeLength)
		if codeErr != nil {
			return store.Link{}, false, codeErr
		}

		link := store.Link{
			Code:      code,
			LongURL:   normalizedURL,
			ExpiresAt: normalizedExpiry,
		}
		saveErr := s.store.Save(ctx, link)
		if saveErr == nil {
			return link, true, nil
		}

		switch {
		case errors.Is(saveErr, store.ErrCodeExists):
			continue
		case errors.Is(saveErr, store.ErrIntentExists):
			existing, ok, findErr := s.store.FindByIntent(ctx, normalizedURL, normalizedExpiry)
			if findErr != nil {
				return store.Link{}, false, findErr
			}
			if ok {
				return existing, false, nil
			}
			return store.Link{}, false, ErrCollision
		default:
			return store.Link{}, false, saveErr
		}
	}

	return store.Link{}, false, ErrCollision
}

func (s *Service) Resolve(ctx context.Context, code string) (store.Link, error) {
	link, ok, err := s.store.Get(ctx, code)
	if err != nil {
		return store.Link{}, err
	}
	if !ok {
		return store.Link{}, ErrNotFound
	}

	if link.ExpiresAt != nil && !link.ExpiresAt.After(s.now().UTC()) {
		return store.Link{}, ErrExpired
	}

	return link, nil
}

func (s *Service) normalizeExpiry(expiresAt *time.Time) (*time.Time, error) {
	if expiresAt == nil {
		return nil, nil
	}

	normalized := expiresAt.UTC()
	if !normalized.After(s.now().UTC()) {
		return nil, ErrExpired
	}
	return &normalized, nil
}
