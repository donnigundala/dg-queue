package queue

import (
	"testing"
	"time"
)

func TestConfig_DefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Driver != "memory" {
		t.Errorf("Expected driver 'memory', got %s", cfg.Driver)
	}
	if cfg.DefaultQueue != "default" {
		t.Errorf("Expected default queue 'default', got %s", cfg.DefaultQueue)
	}
	if cfg.MaxAttempts != 3 {
		t.Errorf("Expected MaxAttempts 3, got %d", cfg.MaxAttempts)
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("Expected Timeout 30s, got %v", cfg.Timeout)
	}
	if cfg.RetryDelay != 1*time.Second {
		t.Errorf("Expected RetryDelay 1s, got %v", cfg.RetryDelay)
	}
	if cfg.Workers != 5 {
		t.Errorf("Expected Workers 5, got %d", cfg.Workers)
	}
}

func TestBatchConfig_DefaultBatchConfig(t *testing.T) {
	cfg := DefaultBatchConfig()

	if cfg.ChunkSize != 100 {
		t.Errorf("Expected ChunkSize 100, got %d", cfg.ChunkSize)
	}
	if cfg.RateLimit != 0 {
		t.Errorf("Expected RateLimit 0, got %d", cfg.RateLimit)
	}
	if cfg.ContinueOnError != true {
		t.Error("Expected ContinueOnError to be true")
	}
}
