package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Config struct {
	ListenAddr       string   `json:"listen_addr"`
	AllowlistedRoots []string `json:"allowlisted_roots"`
	EnvAllowlist     []string `json:"env_allowlist"`
	MaxRuntimeMs     int      `json:"max_runtime_ms"`
	RegistryDir      string   `json:"registry_dir"`
	RunsDir          string   `json:"runs_dir"`
}

func expandHome(p string) string {
	if len(p) > 1 && p[:2] == "~/" {
		h, _ := os.UserHomeDir()
		return filepath.Join(h, p[2:])
	}
	return p
}

func Default() Config {
	return Config{
		ListenAddr:       "127.0.0.1:18789",
		AllowlistedRoots: []string{},
		EnvAllowlist:     []string{"PATH", "HOME", "USER", "SHELL", "TERM"},
		MaxRuntimeMs:     600000,
		RegistryDir:      "~/.musketeer/registry",
		RunsDir:          "~/.musketeer/runs",
	}
}

func Load() (Config, error) {
	cfg := Default()
	h, _ := os.UserHomeDir()
	p := filepath.Join(h, ".musketeer", "bridge.json")
	if b, err := os.ReadFile(p); err == nil {
		if err := json.Unmarshal(b, &cfg); err != nil {
			return cfg, errors.New("ERR_CONFIG_INVALID")
		}
	}
	if v := os.Getenv("MUSKETEER_BRIDGE_LISTEN_ADDR"); v != "" {
		cfg.ListenAddr = v
	}
	if v := os.Getenv("MUSKETEER_BRIDGE_REGISTRY_DIR"); v != "" {
		cfg.RegistryDir = v
	}
	if v := os.Getenv("MUSKETEER_BRIDGE_RUNS_DIR"); v != "" {
		cfg.RunsDir = v
	}
	cfg.RegistryDir = expandHome(cfg.RegistryDir)
	cfg.RunsDir = expandHome(cfg.RunsDir)
	roots := make([]string, 0, len(cfg.AllowlistedRoots))
	for _, r := range cfg.AllowlistedRoots {
		roots = append(roots, expandHome(r))
	}
	cfg.AllowlistedRoots = roots
	return cfg, nil
}
