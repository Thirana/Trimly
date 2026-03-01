package shortener

import (
	"context"
	"errors"
	"sync/atomic"
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
	defaultCacheTTL    = 10 * time.Minute
	defaultMissTTL     = 45 * time.Second
)

type Service struct {
	store       store.LinkStore
	cache       ResolveCache
	cacheTTL    time.Duration
	missTTL     time.Duration
	metrics     resolveMetricsCounter
	codeLength  int
	maxAttempts int
	newCode     func(int) (string, error)
	now         func() time.Time
}

func NewService(s store.LinkStore) *Service {
	return &Service{
		store:       s,
		cache:       noopResolveCache{},
		cacheTTL:    defaultCacheTTL,
		missTTL:     defaultMissTTL,
		codeLength:  defaultCodeLength,
		maxAttempts: defaultMaxAttempts,
		newCode:     NewCode,
		now:         time.Now,
	}
}

func (s *Service) SetResolveCache(cache ResolveCache, cacheTTL time.Duration, missTTL time.Duration) {
	if cache == nil {
		cache = noopResolveCache{}
	}
	if cacheTTL <= 0 {
		cacheTTL = defaultCacheTTL
	}
	if missTTL <= 0 {
		missTTL = defaultMissTTL
	}

	s.cache = cache
	s.cacheTTL = cacheTTL
	s.missTTL = missTTL
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
		s.cacheActiveLink(ctx, existing)
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
			s.cacheActiveLink(ctx, link)
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
				s.cacheActiveLink(ctx, existing)
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
	link, shortHit, err := s.cache.GetShort(ctx, code)
	if err != nil {
		s.metrics.cacheErrors.Add(1)
	}
	if err == nil && shortHit {
		s.metrics.cacheShortHits.Add(1)
		if s.isExpired(link.ExpiresAt) {
			s.cacheDeleteShort(ctx, code)
			s.cacheSetMiss(ctx, code)
			return store.Link{}, ErrExpired
		}
		return link, nil
	}

	missHit, err := s.cache.HasMiss(ctx, code)
	if err != nil {
		s.metrics.cacheErrors.Add(1)
	}
	if err == nil && missHit {
		s.metrics.cacheMissHits.Add(1)
		return store.Link{}, ErrNotFound
	}

	s.metrics.dbFallbacks.Add(1)
	link, ok, err := s.store.Get(ctx, code)
	if err != nil {
		return store.Link{}, err
	}
	if !ok {
		s.metrics.dbMisses.Add(1)
		s.cacheSetMiss(ctx, code)
		return store.Link{}, ErrNotFound
	}

	if s.isExpired(link.ExpiresAt) {
		s.metrics.dbHits.Add(1)
		s.cacheDeleteShort(ctx, code)
		s.cacheSetMiss(ctx, code)
		return store.Link{}, ErrExpired
	}

	s.metrics.dbHits.Add(1)
	s.cacheDeleteMiss(ctx, code)
	if ttl := s.cacheTTLFor(link.ExpiresAt); ttl > 0 {
		s.cacheSetShort(ctx, link, ttl)
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

func (s *Service) isExpired(expiresAt *time.Time) bool {
	if expiresAt == nil {
		return false
	}
	return !expiresAt.After(s.now().UTC())
}

func (s *Service) cacheTTLFor(expiresAt *time.Time) time.Duration {
	if expiresAt == nil {
		return s.cacheTTL
	}

	untilExpiry := expiresAt.UTC().Sub(s.now().UTC())
	if untilExpiry <= 0 {
		return 0
	}
	if untilExpiry < s.cacheTTL {
		return untilExpiry
	}
	return s.cacheTTL
}

func (s *Service) cacheActiveLink(ctx context.Context, link store.Link) {
	s.cacheDeleteMiss(ctx, link.Code)
	if ttl := s.cacheTTLFor(link.ExpiresAt); ttl > 0 {
		s.cacheSetShort(ctx, link, ttl)
	}
}

func (s *Service) cacheSetShort(ctx context.Context, link store.Link, ttl time.Duration) {
	if err := s.cache.SetShort(ctx, link, ttl); err != nil {
		s.metrics.cacheErrors.Add(1)
	}
}

func (s *Service) cacheSetMiss(ctx context.Context, code string) {
	if err := s.cache.SetMiss(ctx, code, s.missTTL); err != nil {
		s.metrics.cacheErrors.Add(1)
	}
}

func (s *Service) cacheDeleteShort(ctx context.Context, code string) {
	if err := s.cache.DeleteShort(ctx, code); err != nil {
		s.metrics.cacheErrors.Add(1)
	}
}

func (s *Service) cacheDeleteMiss(ctx context.Context, code string) {
	if err := s.cache.DeleteMiss(ctx, code); err != nil {
		s.metrics.cacheErrors.Add(1)
	}
}

type ResolveMetricsSnapshot struct {
	CacheShortHits uint64
	CacheMissHits  uint64
	DBFallbacks    uint64
	DBHits         uint64
	DBMisses       uint64
	CacheErrors    uint64
}

type resolveMetricsCounter struct {
	cacheShortHits atomic.Uint64
	cacheMissHits  atomic.Uint64
	dbFallbacks    atomic.Uint64
	dbHits         atomic.Uint64
	dbMisses       atomic.Uint64
	cacheErrors    atomic.Uint64
}

func (s *Service) ResolveMetricsSnapshot() ResolveMetricsSnapshot {
	return ResolveMetricsSnapshot{
		CacheShortHits: s.metrics.cacheShortHits.Load(),
		CacheMissHits:  s.metrics.cacheMissHits.Load(),
		DBFallbacks:    s.metrics.dbFallbacks.Load(),
		DBHits:         s.metrics.dbHits.Load(),
		DBMisses:       s.metrics.dbMisses.Load(),
		CacheErrors:    s.metrics.cacheErrors.Load(),
	}
}
