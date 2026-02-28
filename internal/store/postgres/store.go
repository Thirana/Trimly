package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thirana/url-shortener/internal/store"
)

type Store struct {
	pool *pgxpool.Pool
}

var _ store.LinkStore = (*Store)(nil)

func New(ctx context.Context, dsn string) (*Store, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

func (s *Store) Close() {
	s.pool.Close()
}

func (s *Store) Save(ctx context.Context, link store.Link) error {
	_, err := s.pool.Exec(
		ctx,
		`INSERT INTO links (code, original_url, expires_at, intent_key)
		 VALUES ($1, $2, $3, $4)`,
		link.Code,
		link.LongURL,
		normalizeTime(link.ExpiresAt),
		store.BuildIntentKey(link.LongURL, link.ExpiresAt),
	)
	if err != nil {
		return mapWriteError(err)
	}
	return nil
}

func (s *Store) Get(ctx context.Context, code string) (store.Link, bool, error) {
	var link store.Link
	var expiresAt *time.Time

	err := s.pool.QueryRow(
		ctx,
		`SELECT code, original_url, expires_at FROM links WHERE code = $1`,
		code,
	).Scan(&link.Code, &link.LongURL, &expiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return store.Link{}, false, nil
		}
		return store.Link{}, false, err
	}

	link.ExpiresAt = normalizeTime(expiresAt)
	return link, true, nil
}

func (s *Store) FindByIntent(ctx context.Context, longURL string, expiresAt *time.Time) (store.Link, bool, error) {
	var link store.Link
	var foundExpiresAt *time.Time

	err := s.pool.QueryRow(
		ctx,
		`SELECT code, original_url, expires_at
		 FROM links
		 WHERE intent_key = $1`,
		store.BuildIntentKey(longURL, expiresAt),
	).Scan(&link.Code, &link.LongURL, &foundExpiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return store.Link{}, false, nil
		}
		return store.Link{}, false, err
	}

	link.ExpiresAt = normalizeTime(foundExpiresAt)
	return link, true, nil
}

func mapWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		switch pgErr.ConstraintName {
		case "links_code_key":
			return store.ErrCodeExists
		case "links_intent_key_key":
			return store.ErrIntentExists
		}
	}
	return err
}

func normalizeTime(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	v := t.UTC()
	return &v
}
