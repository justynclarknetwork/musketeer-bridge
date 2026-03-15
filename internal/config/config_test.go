package config_test

import (
	"testing"

	"musketeer-bridge/internal/config"
)

func TestDefaultListenAddr(t *testing.T) {
	cfg := config.Default()
	if cfg.ListenAddr != "127.0.0.1:18789" {
		t.Fatalf("expected 127.0.0.1:18789, got %q", cfg.ListenAddr)
	}
}

func TestDefaultAllowlistEmpty(t *testing.T) {
	cfg := config.Default()
	if len(cfg.AllowlistedRoots) != 0 {
		t.Fatalf("expected empty allowlisted_roots by default, got %v", cfg.AllowlistedRoots)
	}
}

func TestDefaultMaxRuntimeMs(t *testing.T) {
	cfg := config.Default()
	if cfg.MaxRuntimeMs != 600000 {
		t.Fatalf("expected max_runtime_ms 600000, got %d", cfg.MaxRuntimeMs)
	}
}

func TestDefaultEnvAllowlist(t *testing.T) {
	cfg := config.Default()
	if len(cfg.EnvAllowlist) == 0 {
		t.Fatal("expected non-empty env_allowlist by default")
	}
	found := false
	for _, e := range cfg.EnvAllowlist {
		if e == "PATH" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected PATH in env_allowlist, got %v", cfg.EnvAllowlist)
	}
}

func TestListenAddrEnvOverride(t *testing.T) {
	t.Setenv("MUSKETEER_BRIDGE_LISTEN_ADDR", "127.0.0.1:9988")
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ListenAddr != "127.0.0.1:9988" {
		t.Fatalf("expected 127.0.0.1:9988, got %q", cfg.ListenAddr)
	}
}

func TestRegistryDirEnvOverride(t *testing.T) {
	t.Setenv("MUSKETEER_BRIDGE_REGISTRY_DIR", "/tmp/test-registry")
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.RegistryDir != "/tmp/test-registry" {
		t.Fatalf("expected /tmp/test-registry, got %q", cfg.RegistryDir)
	}
}

func TestRunsDirEnvOverride(t *testing.T) {
	t.Setenv("MUSKETEER_BRIDGE_RUNS_DIR", "/tmp/test-runs")
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.RunsDir != "/tmp/test-runs" {
		t.Fatalf("expected /tmp/test-runs, got %q", cfg.RunsDir)
	}
}
