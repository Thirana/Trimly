package shortener

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/thirana/url-shortener/internal/store"
)

func TestNormalizeHTTPURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "lowercases scheme host and adds root path",
			input: "HTTPS://Example.com",
			want:  "https://example.com/",
		},
		{
			name:  "drops default http port",
			input: "http://example.com:80/path",
			want:  "http://example.com/path",
		},
		{
			name:  "drops default https port",
			input: "https://example.com:443",
			want:  "https://example.com/",
		},
		{
			name:    "rejects invalid scheme",
			input:   "ftp://example.com/file",
			wantErr: true,
		},
		{
			name:    "rejects missing host",
			input:   "https:///path",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := NormalizeHTTPURL(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q", tc.input)
				}
				return
			}

			if err != nil {
				t.Fatalf("NormalizeHTTPURL(%q) error: %v", tc.input, err)
			}
			if got != tc.want {
				t.Fatalf("NormalizeHTTPURL(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestServiceCreate_IdempotentByIntent(t *testing.T) {
	t.Parallel()

	mem := store.NewMemoryStore()
	svc := NewService(mem)

	calls := 0
	svc.newCode = func(_ int) (string, error) {
		calls++
		return "sameCode", nil
	}

	ctx := context.Background()
	first, created, err := svc.Create(ctx, "https://Example.com:443", nil)
	if err != nil {
		t.Fatalf("first create returned error: %v", err)
	}
	if !created {
		t.Fatalf("first create should be created=true")
	}

	second, created, err := svc.Create(ctx, "https://example.com/", nil)
	if err != nil {
		t.Fatalf("second create returned error: %v", err)
	}
	if created {
		t.Fatalf("second create should be idempotent created=false")
	}

	if first.Code != second.Code {
		t.Fatalf("idempotent create should return same code: got %q and %q", first.Code, second.Code)
	}
	if calls != 1 {
		t.Fatalf("code generator should be called once, got %d", calls)
	}
}

func TestServiceCreate_RetriesOnCollision(t *testing.T) {
	t.Parallel()

	mem := store.NewMemoryStore()
	ctx := context.Background()

	if err := mem.Save(ctx, store.Link{Code: "dupCode", LongURL: "https://seed.example/"}); err != nil {
		t.Fatalf("seed save failed: %v", err)
	}

	svc := NewService(mem)
	generated := []string{"dupCode", "dupCode", "finalCode"}
	idx := 0
	svc.newCode = func(_ int) (string, error) {
		if idx >= len(generated) {
			return "", errors.New("generator exhausted")
		}
		code := generated[idx]
		idx++
		return code, nil
	}

	link, created, err := svc.Create(ctx, "https://example.com/collision", nil)
	if err != nil {
		t.Fatalf("create returned error: %v", err)
	}
	if !created {
		t.Fatalf("expected created=true after collision retry")
	}
	if link.Code != "finalCode" {
		t.Fatalf("expected finalCode after retries, got %q", link.Code)
	}
}

func TestServiceCreate_ReturnsErrCollisionWhenRetriesExhausted(t *testing.T) {
	t.Parallel()

	mem := store.NewMemoryStore()
	ctx := context.Background()

	if err := mem.Save(ctx, store.Link{Code: "dupCode", LongURL: "https://seed.example/"}); err != nil {
		t.Fatalf("seed save failed: %v", err)
	}

	svc := NewService(mem)
	svc.maxAttempts = 2
	svc.newCode = func(_ int) (string, error) {
		return "dupCode", nil
	}

	_, created, err := svc.Create(ctx, "https://example.com/new", nil)
	if !errors.Is(err, ErrCollision) {
		t.Fatalf("expected ErrCollision, got: %v", err)
	}
	if created {
		t.Fatalf("expected created=false when retries are exhausted")
	}
}

func TestServiceCreate_RejectsExpiredInput(t *testing.T) {
	t.Parallel()

	mem := store.NewMemoryStore()
	svc := NewService(mem)

	now := time.Date(2026, 2, 26, 10, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return now }

	expiresAt := now.Add(-time.Minute)
	_, _, err := svc.Create(context.Background(), "https://example.com/", &expiresAt)
	if !errors.Is(err, ErrExpired) {
		t.Fatalf("expected ErrExpired, got: %v", err)
	}
}

func TestServiceResolve_ReturnsErrExpired(t *testing.T) {
	t.Parallel()

	mem := store.NewMemoryStore()
	svc := NewService(mem)

	now := time.Date(2026, 2, 26, 10, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return now }

	expiresAt := now.Add(-time.Second)
	ctx := context.Background()
	if err := mem.Save(ctx, store.Link{Code: "expired1", LongURL: "https://example.com/", ExpiresAt: &expiresAt}); err != nil {
		t.Fatalf("seed save failed: %v", err)
	}

	_, err := svc.Resolve(ctx, "expired1")
	if !errors.Is(err, ErrExpired) {
		t.Fatalf("expected ErrExpired, got: %v", err)
	}
}

type resolveOnlyStore struct {
	link     store.Link
	ok       bool
	err      error
	getCalls int
}

func (s *resolveOnlyStore) Save(context.Context, store.Link) error {
	return nil
}

func (s *resolveOnlyStore) Get(context.Context, string) (store.Link, bool, error) {
	s.getCalls++
	return s.link, s.ok, s.err
}

func (s *resolveOnlyStore) FindByIntent(context.Context, string, *time.Time) (store.Link, bool, error) {
	return store.Link{}, false, nil
}

type fakeResolveCache struct {
	shortLink store.Link
	shortHit  bool
	shortErr  error

	missHit bool
	missErr error

	setShortCalls int
	setShortLink  store.Link
	setShortTTL   time.Duration
	setShortErr   error

	setMissCalls int
	setMissCode  string
	setMissTTL   time.Duration
	setMissErr   error

	deleteShortCalls int
	deleteShortCode  string
	deleteShortErr   error

	deleteMissCalls int
	deleteMissCode  string
	deleteMissErr   error
}

func (c *fakeResolveCache) GetShort(context.Context, string) (store.Link, bool, error) {
	return c.shortLink, c.shortHit, c.shortErr
}

func (c *fakeResolveCache) HasMiss(context.Context, string) (bool, error) {
	return c.missHit, c.missErr
}

func (c *fakeResolveCache) SetShort(_ context.Context, link store.Link, ttl time.Duration) error {
	c.setShortCalls++
	c.setShortLink = link
	c.setShortTTL = ttl
	return c.setShortErr
}

func (c *fakeResolveCache) SetMiss(_ context.Context, code string, ttl time.Duration) error {
	c.setMissCalls++
	c.setMissCode = code
	c.setMissTTL = ttl
	return c.setMissErr
}

func (c *fakeResolveCache) DeleteShort(_ context.Context, code string) error {
	c.deleteShortCalls++
	c.deleteShortCode = code
	return c.deleteShortErr
}

func (c *fakeResolveCache) DeleteMiss(_ context.Context, code string) error {
	c.deleteMissCalls++
	c.deleteMissCode = code
	return c.deleteMissErr
}

func TestServiceResolve_CacheShortHitSkipsStore(t *testing.T) {
	t.Parallel()

	storeStub := &resolveOnlyStore{
		err: errors.New("store should not be called"),
	}
	cache := &fakeResolveCache{
		shortHit: true,
		shortLink: store.Link{
			Code:    "cache1",
			LongURL: "https://example.com/from-cache",
		},
	}

	svc := NewService(storeStub)
	svc.SetResolveCache(cache, 10*time.Minute, 45*time.Second)

	link, err := svc.Resolve(context.Background(), "cache1")
	if err != nil {
		t.Fatalf("resolve returned error: %v", err)
	}
	if link.LongURL != "https://example.com/from-cache" {
		t.Fatalf("longURL = %q, want %q", link.LongURL, "https://example.com/from-cache")
	}
	if storeStub.getCalls != 0 {
		t.Fatalf("store get calls = %d, want 0", storeStub.getCalls)
	}
}

func TestServiceResolve_CacheMissMarkerSkipsStore(t *testing.T) {
	t.Parallel()

	storeStub := &resolveOnlyStore{
		err: errors.New("store should not be called"),
	}
	cache := &fakeResolveCache{
		shortHit: false,
		missHit:  true,
	}

	svc := NewService(storeStub)
	svc.SetResolveCache(cache, 10*time.Minute, 45*time.Second)

	_, err := svc.Resolve(context.Background(), "missing1")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if storeStub.getCalls != 0 {
		t.Fatalf("store get calls = %d, want 0", storeStub.getCalls)
	}
}

func TestServiceResolve_DBMissSetsNegativeCache(t *testing.T) {
	t.Parallel()

	storeStub := &resolveOnlyStore{
		ok: false,
	}
	cache := &fakeResolveCache{}

	svc := NewService(storeStub)
	svc.SetResolveCache(cache, 10*time.Minute, 45*time.Second)

	_, err := svc.Resolve(context.Background(), "missing2")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if cache.setMissCalls != 1 {
		t.Fatalf("setMiss calls = %d, want 1", cache.setMissCalls)
	}
	if cache.setMissCode != "missing2" {
		t.Fatalf("setMiss code = %q, want %q", cache.setMissCode, "missing2")
	}
}

func TestServiceResolve_DBHitSetsPositiveCache(t *testing.T) {
	t.Parallel()

	storeStub := &resolveOnlyStore{
		ok: true,
		link: store.Link{
			Code:    "dbhit1",
			LongURL: "https://example.com/from-db",
		},
	}
	cache := &fakeResolveCache{}

	svc := NewService(storeStub)
	svc.SetResolveCache(cache, 10*time.Minute, 45*time.Second)

	link, err := svc.Resolve(context.Background(), "dbhit1")
	if err != nil {
		t.Fatalf("resolve returned error: %v", err)
	}
	if link.LongURL != "https://example.com/from-db" {
		t.Fatalf("longURL = %q, want %q", link.LongURL, "https://example.com/from-db")
	}
	if cache.setShortCalls != 1 {
		t.Fatalf("setShort calls = %d, want 1", cache.setShortCalls)
	}
	if cache.deleteMissCalls != 1 {
		t.Fatalf("deleteMiss calls = %d, want 1", cache.deleteMissCalls)
	}
	if cache.setShortTTL != 10*time.Minute {
		t.Fatalf("setShort ttl = %v, want %v", cache.setShortTTL, 10*time.Minute)
	}
}

func TestServiceResolve_PositiveTTLClampedToExpiry(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 1, 11, 0, 0, 0, time.UTC)
	expiresAt := now.Add(30 * time.Second)

	storeStub := &resolveOnlyStore{
		ok: true,
		link: store.Link{
			Code:      "dbhit2",
			LongURL:   "https://example.com/expiring-soon",
			ExpiresAt: &expiresAt,
		},
	}
	cache := &fakeResolveCache{}

	svc := NewService(storeStub)
	svc.now = func() time.Time { return now }
	svc.SetResolveCache(cache, 10*time.Minute, 45*time.Second)

	_, err := svc.Resolve(context.Background(), "dbhit2")
	if err != nil {
		t.Fatalf("resolve returned error: %v", err)
	}
	if cache.setShortTTL != 30*time.Second {
		t.Fatalf("setShort ttl = %v, want %v", cache.setShortTTL, 30*time.Second)
	}
}

func TestServiceCreate_ClearsMissAndWarmsPositiveCache(t *testing.T) {
	t.Parallel()

	mem := store.NewMemoryStore()
	cache := &fakeResolveCache{}
	svc := NewService(mem)
	svc.SetResolveCache(cache, 10*time.Minute, 45*time.Second)
	svc.newCode = func(_ int) (string, error) {
		return "newcache1", nil
	}

	link, created, err := svc.Create(context.Background(), "https://example.com/new", nil)
	if err != nil {
		t.Fatalf("create returned error: %v", err)
	}
	if !created {
		t.Fatalf("expected created=true")
	}
	if cache.deleteMissCalls != 1 {
		t.Fatalf("deleteMiss calls = %d, want 1", cache.deleteMissCalls)
	}
	if cache.deleteMissCode != link.Code {
		t.Fatalf("deleteMiss code = %q, want %q", cache.deleteMissCode, link.Code)
	}
	if cache.setShortCalls != 1 {
		t.Fatalf("setShort calls = %d, want 1", cache.setShortCalls)
	}
	if cache.setShortTTL != 10*time.Minute {
		t.Fatalf("setShort ttl = %v, want %v", cache.setShortTTL, 10*time.Minute)
	}
}

func TestServiceResolveMetrics_CacheShortHit(t *testing.T) {
	t.Parallel()

	storeStub := &resolveOnlyStore{
		err: errors.New("store should not be called"),
	}
	cache := &fakeResolveCache{
		shortHit: true,
		shortLink: store.Link{
			Code:    "cachemetrics1",
			LongURL: "https://example.com/cache",
		},
	}

	svc := NewService(storeStub)
	svc.SetResolveCache(cache, 10*time.Minute, 45*time.Second)

	if _, err := svc.Resolve(context.Background(), "cachemetrics1"); err != nil {
		t.Fatalf("resolve returned error: %v", err)
	}

	m := svc.ResolveMetricsSnapshot()
	if m.CacheShortHits != 1 {
		t.Fatalf("CacheShortHits = %d, want 1", m.CacheShortHits)
	}
	if m.DBFallbacks != 0 {
		t.Fatalf("DBFallbacks = %d, want 0", m.DBFallbacks)
	}
}

func TestServiceResolveMetrics_FallbackAndCacheErrors(t *testing.T) {
	t.Parallel()

	storeStub := &resolveOnlyStore{
		ok: false,
	}
	cache := &fakeResolveCache{
		shortErr: errors.New("redis short get failed"),
		missErr:  errors.New("redis miss get failed"),
	}

	svc := NewService(storeStub)
	svc.SetResolveCache(cache, 10*time.Minute, 45*time.Second)

	_, err := svc.Resolve(context.Background(), "cachemetrics2")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	m := svc.ResolveMetricsSnapshot()
	if m.CacheErrors < 2 {
		t.Fatalf("CacheErrors = %d, want at least 2", m.CacheErrors)
	}
	if m.DBFallbacks != 1 {
		t.Fatalf("DBFallbacks = %d, want 1", m.DBFallbacks)
	}
	if m.DBMisses != 1 {
		t.Fatalf("DBMisses = %d, want 1", m.DBMisses)
	}
}
