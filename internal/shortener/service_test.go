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
