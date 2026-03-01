package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/thirana/url-shortener/internal/store"
)

const (
	migrationUpFile   = "000001_create_links_table.up.sql"
	migrationDownFile = "000001_create_links_table.down.sql"
)

func TestStore_SaveAndGet_RoundTrip(t *testing.T) {
	s := newIntegrationStore(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	expiresAt := time.Date(2026, 3, 2, 10, 0, 0, 0, time.FixedZone("LCL", 5*60*60))
	link := store.Link{
		Code:      newUniqueCode("rt"),
		LongURL:   "https://example.com/round-trip",
		ExpiresAt: &expiresAt,
	}

	if err := s.Save(ctx, link); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	got, ok, err := s.Get(ctx, link.Code)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if !ok {
		t.Fatalf("get ok=false, want true")
	}
	if got.Code != link.Code {
		t.Fatalf("got code %q, want %q", got.Code, link.Code)
	}
	if got.LongURL != link.LongURL {
		t.Fatalf("got long_url %q, want %q", got.LongURL, link.LongURL)
	}
	if got.ExpiresAt == nil {
		t.Fatalf("expires_at is nil, want value")
	}
	if !got.ExpiresAt.Equal(expiresAt.UTC()) {
		t.Fatalf("expires_at = %v, want %v", got.ExpiresAt.UTC(), expiresAt.UTC())
	}
}

func TestStore_Get_NotFound(t *testing.T) {
	s := newIntegrationStore(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, ok, err := s.Get(ctx, newUniqueCode("missing"))
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if ok {
		t.Fatalf("get ok=true, want false")
	}
}

func TestStore_FindByIntent(t *testing.T) {
	s := newIntegrationStore(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	expiresAtUTC := time.Date(2026, 3, 3, 9, 0, 0, 0, time.UTC)
	expiresAtLocal := expiresAtUTC.In(time.FixedZone("LCL", 7*60*60))

	link := store.Link{
		Code:      newUniqueCode("intent"),
		LongURL:   "https://example.com/by-intent",
		ExpiresAt: &expiresAtLocal,
	}
	if err := s.Save(ctx, link); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	got, ok, err := s.FindByIntent(ctx, link.LongURL, &expiresAtUTC)
	if err != nil {
		t.Fatalf("find by intent failed: %v", err)
	}
	if !ok {
		t.Fatalf("find by intent ok=false, want true")
	}
	if got.Code != link.Code {
		t.Fatalf("got code %q, want %q", got.Code, link.Code)
	}
	if got.ExpiresAt == nil || !got.ExpiresAt.Equal(expiresAtUTC) {
		t.Fatalf("got expires_at %v, want %v", got.ExpiresAt, expiresAtUTC)
	}

	_, ok, err = s.FindByIntent(ctx, "https://example.com/does-not-exist", nil)
	if err != nil {
		t.Fatalf("find by intent (missing) failed: %v", err)
	}
	if ok {
		t.Fatalf("find by intent (missing) ok=true, want false")
	}
}

func TestStore_Save_UniqueConstraintMapping(t *testing.T) {
	s := newIntegrationStore(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sharedCode := newUniqueCode("dupcode")
	firstCode := store.Link{
		Code:    sharedCode,
		LongURL: "https://example.com/first",
	}
	if err := s.Save(ctx, firstCode); err != nil {
		t.Fatalf("save first code failed: %v", err)
	}

	secondCode := store.Link{
		Code:    sharedCode,
		LongURL: "https://example.com/second",
	}
	if err := s.Save(ctx, secondCode); !errors.Is(err, store.ErrCodeExists) {
		t.Fatalf("duplicate code error = %v, want %v", err, store.ErrCodeExists)
	}

	intentURL := "https://example.com/shared-intent"
	firstIntent := store.Link{
		Code:    newUniqueCode("intent-a"),
		LongURL: intentURL,
	}
	if err := s.Save(ctx, firstIntent); err != nil {
		t.Fatalf("save first intent failed: %v", err)
	}

	secondIntent := store.Link{
		Code:    newUniqueCode("intent-b"),
		LongURL: intentURL,
	}
	if err := s.Save(ctx, secondIntent); !errors.Is(err, store.ErrIntentExists) {
		t.Fatalf("duplicate intent error = %v, want %v", err, store.ErrIntentExists)
	}
}

func TestMigrations_UpDownUp(t *testing.T) {
	env := newTestSchemaEnvironment(t)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := applyMigration(ctx, env.adminPool, env.schema, migrationUpFile); err != nil {
		t.Fatalf("migration up failed: %v", err)
	}
	assertLinksTableExists(t, ctx, env.adminPool, env.schema, true)
	assertUniqueConstraintsPresent(t, ctx, env.adminPool, env.schema)

	if err := applyMigration(ctx, env.adminPool, env.schema, migrationDownFile); err != nil {
		t.Fatalf("migration down failed: %v", err)
	}
	assertLinksTableExists(t, ctx, env.adminPool, env.schema, false)

	if err := applyMigration(ctx, env.adminPool, env.schema, migrationUpFile); err != nil {
		t.Fatalf("migration up (second pass) failed: %v", err)
	}
	assertLinksTableExists(t, ctx, env.adminPool, env.schema, true)
}

type testSchemaEnvironment struct {
	adminPool *pgxpool.Pool
	schema    string
	dsn       string
}

func newIntegrationStore(t *testing.T) *Store {
	t.Helper()

	env := newTestSchemaEnvironment(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := applyMigration(ctx, env.adminPool, env.schema, migrationUpFile); err != nil {
		t.Fatalf("migration up failed: %v", err)
	}

	cfg, err := pgxpool.ParseConfig(env.dsn)
	if err != nil {
		t.Fatalf("parse config failed: %v", err)
	}
	cfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, "SET search_path TO "+env.schema)
		return err
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("create store pool failed: %v", err)
	}

	store := &Store{pool: pool}
	t.Cleanup(store.Close)
	return store
}

func newTestSchemaEnvironment(t *testing.T) testSchemaEnvironment {
	t.Helper()

	dsn := integrationDatabaseURL(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	adminPool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("create admin pool failed: %v", err)
	}

	schema := newSchemaName(t.Name())
	if _, err := adminPool.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		adminPool.Close()
		t.Fatalf("create schema failed: %v", err)
	}

	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		_, err := adminPool.Exec(cleanupCtx, "DROP SCHEMA "+schema+" CASCADE")
		adminPool.Close()
		if err != nil {
			t.Fatalf("drop schema failed: %v", err)
		}
	})

	return testSchemaEnvironment{
		adminPool: adminPool,
		schema:    schema,
		dsn:       dsn,
	}
}

func applyMigration(ctx context.Context, pool *pgxpool.Pool, schema string, migrationFile string) error {
	queryBytes, err := os.ReadFile(migrationPath(migrationFile))
	if err != nil {
		return err
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL search_path TO "+schema); err != nil {
		return err
	}

	for _, stmt := range splitSQLStatements(string(queryBytes)) {
		if _, err := tx.Exec(ctx, stmt); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func assertLinksTableExists(t *testing.T, ctx context.Context, pool *pgxpool.Pool, schema string, want bool) {
	t.Helper()

	var exists bool
	err := pool.QueryRow(
		ctx,
		`SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = $1 AND table_name = 'links'
		)`,
		schema,
	).Scan(&exists)
	if err != nil {
		t.Fatalf("query links table existence failed: %v", err)
	}
	if exists != want {
		t.Fatalf("links table exists=%v, want %v", exists, want)
	}
}

func assertUniqueConstraintsPresent(t *testing.T, ctx context.Context, pool *pgxpool.Pool, schema string) {
	t.Helper()

	rows, err := pool.Query(
		ctx,
		`SELECT constraint_name
		 FROM information_schema.table_constraints
		 WHERE table_schema = $1
		   AND table_name = 'links'
		   AND constraint_type = 'UNIQUE'`,
		schema,
	)
	if err != nil {
		t.Fatalf("query constraints failed: %v", err)
	}
	defer rows.Close()

	names := make(map[string]struct{})
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scan constraint name failed: %v", err)
		}
		names[name] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("read constraints failed: %v", err)
	}

	required := []string{"links_code_key", "links_intent_key_key"}
	for _, name := range required {
		if _, ok := names[name]; !ok {
			t.Fatalf("missing unique constraint %q", name)
		}
	}
}

func splitSQLStatements(sql string) []string {
	parts := strings.Split(sql, ";")
	statements := make([]string, 0, len(parts))
	for _, part := range parts {
		stmt := strings.TrimSpace(part)
		if stmt == "" {
			continue
		}
		statements = append(statements, stmt)
	}
	return statements
}

func migrationPath(fileName string) string {
	_, currentFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(currentFile), "migrations", fileName)
}

func integrationDatabaseURL(t *testing.T) string {
	t.Helper()

	if dsn := strings.TrimSpace(os.Getenv("DATABASE_URL")); dsn != "" {
		return dsn
	}

	loadDotEnvForTests()
	if dsn := strings.TrimSpace(os.Getenv("DATABASE_URL")); dsn != "" {
		return dsn
	}

	t.Skip("DATABASE_URL is not set; skipping postgres integration tests")
	return ""
}

func loadDotEnvForTests() {
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")
	_ = godotenv.Load("../../.env")
	_ = godotenv.Load("../../../.env")
}

func newSchemaName(testName string) string {
	name := strings.ToLower(testName)
	name = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= '0' && r <= '9':
			return r
		default:
			return '_'
		}
	}, name)
	if len(name) > 20 {
		name = name[:20]
	}
	return fmt.Sprintf("it_%s_%d", name, time.Now().UnixNano())
}

func newUniqueCode(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}
