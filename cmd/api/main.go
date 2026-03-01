package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/thirana/url-shortener/internal/httpapi"
	"github.com/thirana/url-shortener/internal/shortener"
	"github.com/thirana/url-shortener/internal/store"
	storepostgres "github.com/thirana/url-shortener/internal/store/postgres"
	"github.com/thirana/url-shortener/internal/store/rediscache"
)

func main() {
	if err := godotenv.Load(".env"); err != nil && !os.IsNotExist(err) {
		log.Printf("warning: failed to load .env: %v", err)
	}

	// Production default: release mode reduces debug noise and overhead.
	// You can also set: export GIN_MODE=release
	gin.SetMode(gin.ReleaseMode)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	redisCfg, err := loadRedisConfig()
	if err != nil {
		log.Fatalf("invalid redis config: %v", err)
	}
	resolveCache, closeRedis, err := buildRedis(redisCfg)
	if err != nil {
		log.Fatalf("failed to initialize redis: %v", err)
	}

	linkStore, closeStore, err := buildStore()
	if err != nil {
		log.Fatalf("failed to initialize store: %v", err)
	}

	svc := shortener.NewService(linkStore)
	svc.SetResolveCache(resolveCache, redisCfg.PositiveTTL, redisCfg.MissTTL)
	stopCacheMetrics := startCacheMetricsLogger(svc, redisCfg.Enabled, redisCfg.MetricsLogInterval)
	links := httpapi.NewLinksHandler(svc)
	router := httpapi.NewRouter(links)

	log.Printf("starting api on :%s", port)
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		log.Fatalf("server failed: %v", err)
	case sig := <-stop:
		log.Printf("shutdown signal received: %s", sig.String())
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}

	if err := closeStore(); err != nil {
		log.Printf("store close failed: %v", err)
	}
	if err := closeRedis(); err != nil {
		log.Printf("redis close failed: %v", err)
	}
	stopCacheMetrics()
}

func buildStore() (store.LinkStore, func() error, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Printf("using in-memory store (set DATABASE_URL to enable Postgres)")
		return store.NewMemoryStore(), func() error { return nil }, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pgStore, err := storepostgres.New(ctx, databaseURL)
	if err != nil {
		return nil, nil, err
	}
	if err := pgStore.Ping(ctx); err != nil {
		pgStore.Close()
		return nil, nil, err
	}

	log.Printf("using postgres store")
	log.Printf("postgres connectivity check passed")
	return pgStore, func() error {
		pgStore.Close()
		return nil
	}, nil
}

type redisConfig struct {
	Enabled            bool
	URL                string
	PositiveTTL        time.Duration
	MissTTL            time.Duration
	ConnectTimeout     time.Duration
	OpTimeout          time.Duration
	MetricsLogInterval time.Duration
}

const (
	defaultRedisPositiveTTL        = 10 * time.Minute
	defaultRedisMissTTL            = 45 * time.Second
	defaultRedisConnectTimeout     = 3 * time.Second
	defaultRedisOpTimeout          = 150 * time.Millisecond
	defaultCacheMetricsLogInterval = 30 * time.Second
)

func loadRedisConfig() (redisConfig, error) {
	enabled, err := boolFromEnv("REDIS_ENABLED", false)
	if err != nil {
		return redisConfig{}, err
	}

	positiveTTL, err := durationFromEnv("REDIS_POSITIVE_TTL", defaultRedisPositiveTTL)
	if err != nil {
		return redisConfig{}, err
	}
	missTTL, err := durationFromEnv("REDIS_MISS_TTL", defaultRedisMissTTL)
	if err != nil {
		return redisConfig{}, err
	}
	connectTimeout, err := durationFromEnv("REDIS_CONNECT_TIMEOUT", defaultRedisConnectTimeout)
	if err != nil {
		return redisConfig{}, err
	}
	opTimeout, err := durationFromEnvWithAlias("REDIS_OP_TIMEOUT", "REDIS_TIMEOUT", defaultRedisOpTimeout)
	if err != nil {
		return redisConfig{}, err
	}
	metricsLogInterval, err := durationFromEnv("CACHE_METRICS_LOG_INTERVAL", defaultCacheMetricsLogInterval)
	if err != nil {
		return redisConfig{}, err
	}

	cfg := redisConfig{
		Enabled:            enabled,
		URL:                strings.TrimSpace(os.Getenv("REDIS_URL")),
		PositiveTTL:        positiveTTL,
		MissTTL:            missTTL,
		ConnectTimeout:     connectTimeout,
		OpTimeout:          opTimeout,
		MetricsLogInterval: metricsLogInterval,
	}

	if cfg.PositiveTTL <= 0 {
		return redisConfig{}, fmt.Errorf("REDIS_POSITIVE_TTL must be > 0")
	}
	if cfg.MissTTL <= 0 {
		return redisConfig{}, fmt.Errorf("REDIS_MISS_TTL must be > 0")
	}
	if cfg.ConnectTimeout <= 0 {
		return redisConfig{}, fmt.Errorf("REDIS_CONNECT_TIMEOUT must be > 0")
	}
	if cfg.OpTimeout <= 0 {
		return redisConfig{}, fmt.Errorf("REDIS_OP_TIMEOUT must be > 0")
	}
	if cfg.MetricsLogInterval < 0 {
		return redisConfig{}, fmt.Errorf("CACHE_METRICS_LOG_INTERVAL must be >= 0")
	}
	if cfg.Enabled && cfg.URL == "" {
		return redisConfig{}, fmt.Errorf("REDIS_URL must be set when REDIS_ENABLED=true")
	}

	return cfg, nil
}

func boolFromEnv(key string, fallback bool) (bool, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean: %w", key, err)
	}
	return v, nil
}

func durationFromEnv(key string, fallback time.Duration) (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	v, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration: %w", key, err)
	}
	return v, nil
}

func durationFromEnvWithAlias(primaryKey, aliasKey string, fallback time.Duration) (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(primaryKey))
	if raw == "" {
		raw = strings.TrimSpace(os.Getenv(aliasKey))
	}
	if raw == "" {
		return fallback, nil
	}
	v, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration: %w", primaryKey, err)
	}
	return v, nil
}

func logRedisConfig(cfg redisConfig) {
	if !cfg.Enabled {
		log.Printf("redis cache disabled")
		return
	}

	log.Printf(
		"redis cache enabled positive_ttl=%s miss_ttl=%s connect_timeout=%s op_timeout=%s metrics_log_interval=%s",
		cfg.PositiveTTL,
		cfg.MissTTL,
		cfg.ConnectTimeout,
		cfg.OpTimeout,
		cfg.MetricsLogInterval,
	)
}

func buildRedis(cfg redisConfig) (shortener.ResolveCache, func() error, error) {
	logRedisConfig(cfg)

	if !cfg.Enabled {
		return nil, func() error { return nil }, nil
	}

	opts, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, nil, fmt.Errorf("parse REDIS_URL: %w", err)
	}
	opts.DialTimeout = cfg.ConnectTimeout
	opts.ReadTimeout = cfg.OpTimeout
	opts.WriteTimeout = cfg.OpTimeout

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnectTimeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, nil, fmt.Errorf("ping redis: %w", err)
	}

	log.Printf("redis connectivity check passed")
	return rediscache.New(client, cfg.OpTimeout), client.Close, nil
}

func startCacheMetricsLogger(svc *shortener.Service, redisEnabled bool, interval time.Duration) func() {
	if !redisEnabled {
		return func() {}
	}
	if interval <= 0 {
		log.Printf("cache metrics logger disabled (CACHE_METRICS_LOG_INTERVAL=%s)", interval)
		return func() {}
	}

	ctx, cancel := context.WithCancel(context.Background())
	ticker := time.NewTicker(interval)
	log.Printf("cache metrics logger enabled interval=%s", interval)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				m := svc.ResolveMetricsSnapshot()
				log.Printf(
					"cache_metrics short_hit=%d miss_hit=%d db_fallback=%d db_hit=%d db_miss=%d cache_error=%d",
					m.CacheShortHits,
					m.CacheMissHits,
					m.DBFallbacks,
					m.DBHits,
					m.DBMisses,
					m.CacheErrors,
				)
			case <-ctx.Done():
				return
			}
		}
	}()

	return cancel
}
