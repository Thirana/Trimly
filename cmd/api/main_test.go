package main

import (
	"testing"
	"time"
)

func TestLoadRedisConfig_Defaults(t *testing.T) {
	t.Setenv("REDIS_ENABLED", "")
	t.Setenv("REDIS_URL", "")
	t.Setenv("REDIS_POSITIVE_TTL", "")
	t.Setenv("REDIS_MISS_TTL", "")
	t.Setenv("REDIS_CONNECT_TIMEOUT", "")
	t.Setenv("REDIS_OP_TIMEOUT", "")
	t.Setenv("REDIS_TIMEOUT", "")
	t.Setenv("CACHE_METRICS_LOG_INTERVAL", "")

	cfg, err := loadRedisConfig()
	if err != nil {
		t.Fatalf("loadRedisConfig returned error: %v", err)
	}

	if cfg.Enabled {
		t.Fatalf("Enabled = true, want false")
	}
	if cfg.PositiveTTL != defaultRedisPositiveTTL {
		t.Fatalf("PositiveTTL = %v, want %v", cfg.PositiveTTL, defaultRedisPositiveTTL)
	}
	if cfg.MissTTL != defaultRedisMissTTL {
		t.Fatalf("MissTTL = %v, want %v", cfg.MissTTL, defaultRedisMissTTL)
	}
	if cfg.ConnectTimeout != defaultRedisConnectTimeout {
		t.Fatalf("ConnectTimeout = %v, want %v", cfg.ConnectTimeout, defaultRedisConnectTimeout)
	}
	if cfg.OpTimeout != defaultRedisOpTimeout {
		t.Fatalf("OpTimeout = %v, want %v", cfg.OpTimeout, defaultRedisOpTimeout)
	}
	if cfg.MetricsLogInterval != defaultCacheMetricsLogInterval {
		t.Fatalf("MetricsLogInterval = %v, want %v", cfg.MetricsLogInterval, defaultCacheMetricsLogInterval)
	}
}

func TestLoadRedisConfig_CustomValues(t *testing.T) {
	t.Setenv("REDIS_ENABLED", "true")
	t.Setenv("REDIS_URL", "rediss://default:pass@localhost:6379")
	t.Setenv("REDIS_POSITIVE_TTL", "15m")
	t.Setenv("REDIS_MISS_TTL", "30s")
	t.Setenv("REDIS_CONNECT_TIMEOUT", "3s")
	t.Setenv("REDIS_OP_TIMEOUT", "200ms")
	t.Setenv("CACHE_METRICS_LOG_INTERVAL", "10s")

	cfg, err := loadRedisConfig()
	if err != nil {
		t.Fatalf("loadRedisConfig returned error: %v", err)
	}

	if !cfg.Enabled {
		t.Fatalf("Enabled = false, want true")
	}
	if cfg.URL == "" {
		t.Fatalf("URL = empty, want set")
	}
	if got, want := cfg.PositiveTTL.String(), "15m0s"; got != want {
		t.Fatalf("PositiveTTL = %s, want %s", got, want)
	}
	if got, want := cfg.MissTTL.String(), "30s"; got != want {
		t.Fatalf("MissTTL = %s, want %s", got, want)
	}
	if got, want := cfg.ConnectTimeout.String(), "3s"; got != want {
		t.Fatalf("ConnectTimeout = %s, want %s", got, want)
	}
	if got, want := cfg.OpTimeout.String(), "200ms"; got != want {
		t.Fatalf("OpTimeout = %s, want %s", got, want)
	}
	if got, want := cfg.MetricsLogInterval.String(), "10s"; got != want {
		t.Fatalf("MetricsLogInterval = %s, want %s", got, want)
	}
}

func TestLoadRedisConfig_LegacyTimeoutFallback(t *testing.T) {
	t.Setenv("REDIS_ENABLED", "true")
	t.Setenv("REDIS_URL", "rediss://default:pass@localhost:6379")
	t.Setenv("REDIS_OP_TIMEOUT", "")
	t.Setenv("REDIS_TIMEOUT", "220ms")

	cfg, err := loadRedisConfig()
	if err != nil {
		t.Fatalf("loadRedisConfig returned error: %v", err)
	}
	if got, want := cfg.OpTimeout.String(), "220ms"; got != want {
		t.Fatalf("OpTimeout = %s, want %s", got, want)
	}
}

func TestLoadRedisConfig_EnabledRequiresURL(t *testing.T) {
	t.Setenv("REDIS_ENABLED", "true")
	t.Setenv("REDIS_URL", "")

	_, err := loadRedisConfig()
	if err == nil {
		t.Fatalf("expected error when REDIS_URL is empty and REDIS_ENABLED=true")
	}
}

func TestLoadRedisConfig_InvalidBool(t *testing.T) {
	t.Setenv("REDIS_ENABLED", "not-a-bool")

	_, err := loadRedisConfig()
	if err == nil {
		t.Fatalf("expected error for invalid REDIS_ENABLED")
	}
}

func TestLoadRedisConfig_InvalidDuration(t *testing.T) {
	t.Setenv("REDIS_ENABLED", "false")
	t.Setenv("REDIS_CONNECT_TIMEOUT", "abc")

	_, err := loadRedisConfig()
	if err == nil {
		t.Fatalf("expected error for invalid REDIS_CONNECT_TIMEOUT")
	}
}

func TestLoadRedisConfig_InvalidMetricsInterval(t *testing.T) {
	t.Setenv("REDIS_ENABLED", "false")
	t.Setenv("CACHE_METRICS_LOG_INTERVAL", "-1s")

	_, err := loadRedisConfig()
	if err == nil {
		t.Fatalf("expected error for invalid CACHE_METRICS_LOG_INTERVAL")
	}
}

func TestBuildRedis_Disabled(t *testing.T) {
	cache, closeFn, err := buildRedis(redisConfig{
		Enabled:        false,
		PositiveTTL:    defaultRedisPositiveTTL,
		MissTTL:        defaultRedisMissTTL,
		ConnectTimeout: defaultRedisConnectTimeout,
		OpTimeout:      defaultRedisOpTimeout,
	})
	if err != nil {
		t.Fatalf("buildRedis returned error: %v", err)
	}
	if cache != nil {
		t.Fatalf("cache = %v, want nil when redis disabled", cache)
	}
	if closeFn == nil {
		t.Fatalf("closeFn is nil")
	}
	if err := closeFn(); err != nil {
		t.Fatalf("closeFn returned error: %v", err)
	}
}

func TestBuildRedis_InvalidURL(t *testing.T) {
	_, _, err := buildRedis(redisConfig{
		Enabled:        true,
		URL:            "not-a-redis-url",
		PositiveTTL:    defaultRedisPositiveTTL,
		MissTTL:        defaultRedisMissTTL,
		ConnectTimeout: 3 * time.Second,
		OpTimeout:      200 * time.Millisecond,
	})
	if err == nil {
		t.Fatalf("expected error for invalid REDIS_URL")
	}
}
