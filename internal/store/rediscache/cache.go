package rediscache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/thirana/url-shortener/internal/store"
)

const (
	shortKeyPrefix = "v1:url:short:"
	missKeyPrefix  = "v1:url:miss:"
	missValue      = "1"
)

type Cache struct {
	client    *redis.Client
	opTimeout time.Duration
}

type cachedShortValue struct {
	LongURL       string `json:"u"`
	ExpiresAtNano *int64 `json:"x,omitempty"`
}

func New(client *redis.Client, opTimeout time.Duration) *Cache {
	return &Cache{
		client:    client,
		opTimeout: opTimeout,
	}
}

func (c *Cache) GetShort(ctx context.Context, code string) (store.Link, bool, error) {
	opCtx, cancel := c.withTimeout(ctx)
	defer cancel()

	val, err := c.client.Get(opCtx, shortKey(code)).Result()
	if err != nil {
		if err == redis.Nil {
			return store.Link{}, false, nil
		}
		return store.Link{}, false, err
	}

	var payload cachedShortValue
	if err := json.Unmarshal([]byte(val), &payload); err != nil {
		return store.Link{}, false, err
	}

	link := store.Link{
		Code:    code,
		LongURL: payload.LongURL,
	}
	if payload.ExpiresAtNano != nil {
		expires := time.Unix(0, *payload.ExpiresAtNano).UTC()
		link.ExpiresAt = &expires
	}
	return link, true, nil
}

func (c *Cache) HasMiss(ctx context.Context, code string) (bool, error) {
	opCtx, cancel := c.withTimeout(ctx)
	defer cancel()

	_, err := c.client.Get(opCtx, missKey(code)).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *Cache) SetShort(ctx context.Context, link store.Link, ttl time.Duration) error {
	if ttl <= 0 {
		return nil
	}
	payload, err := encodeShortValue(link)
	if err != nil {
		return err
	}
	opCtx, cancel := c.withTimeout(ctx)
	defer cancel()

	return c.client.Set(opCtx, shortKey(link.Code), payload, ttl).Err()
}

func (c *Cache) SetMiss(ctx context.Context, code string, ttl time.Duration) error {
	if ttl <= 0 {
		return nil
	}
	opCtx, cancel := c.withTimeout(ctx)
	defer cancel()

	return c.client.Set(opCtx, missKey(code), missValue, ttl).Err()
}

func (c *Cache) DeleteShort(ctx context.Context, code string) error {
	opCtx, cancel := c.withTimeout(ctx)
	defer cancel()

	return c.client.Del(opCtx, shortKey(code)).Err()
}

func (c *Cache) DeleteMiss(ctx context.Context, code string) error {
	opCtx, cancel := c.withTimeout(ctx)
	defer cancel()

	return c.client.Del(opCtx, missKey(code)).Err()
}

func shortKey(code string) string {
	return shortKeyPrefix + code
}

func missKey(code string) string {
	return missKeyPrefix + code
}

func encodeShortValue(link store.Link) (string, error) {
	payload := cachedShortValue{
		LongURL: link.LongURL,
	}
	if link.ExpiresAt != nil {
		v := link.ExpiresAt.UTC().UnixNano()
		payload.ExpiresAtNano = &v
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal short cache payload: %w", err)
	}
	return string(raw), nil
}

func (c *Cache) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if c.opTimeout <= 0 {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, c.opTimeout)
}
