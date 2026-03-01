package shortener

import (
	"context"
	"time"

	"github.com/thirana/url-shortener/internal/store"
)

type ResolveCache interface {
	GetShort(ctx context.Context, code string) (store.Link, bool, error)
	HasMiss(ctx context.Context, code string) (bool, error)
	SetShort(ctx context.Context, link store.Link, ttl time.Duration) error
	SetMiss(ctx context.Context, code string, ttl time.Duration) error
	DeleteShort(ctx context.Context, code string) error
	DeleteMiss(ctx context.Context, code string) error
}

type noopResolveCache struct{}

func (noopResolveCache) GetShort(context.Context, string) (store.Link, bool, error) {
	return store.Link{}, false, nil
}

func (noopResolveCache) HasMiss(context.Context, string) (bool, error) {
	return false, nil
}

func (noopResolveCache) SetShort(context.Context, store.Link, time.Duration) error {
	return nil
}

func (noopResolveCache) SetMiss(context.Context, string, time.Duration) error {
	return nil
}

func (noopResolveCache) DeleteShort(context.Context, string) error {
	return nil
}

func (noopResolveCache) DeleteMiss(context.Context, string) error {
	return nil
}
