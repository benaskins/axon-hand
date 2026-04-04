package agent

import (
	"os"
	"testing"
)

func clearEnv(t *testing.T, keys ...string) {
	t.Helper()
	for _, k := range keys {
		t.Setenv(k, "")
		os.Unsetenv(k)
	}
}

func TestLoadConfig_PrefixVars(t *testing.T) {
	clearEnv(t,
		"INSPECTOR_PROVIDER", "INSPECTOR_MODEL", "INSPECTOR_API_KEY", "INSPECTOR_BASE_URL",
		"FACTORY_PROVIDER", "FACTORY_MODEL", "FACTORY_API_KEY", "FACTORY_BASE_URL",
	)
	t.Setenv("INSPECTOR_PROVIDER", "openrouter")
	t.Setenv("INSPECTOR_MODEL", "qwen/qwen3.5-122b-a10b")
	t.Setenv("INSPECTOR_API_KEY", "sk-test")
	t.Setenv("INSPECTOR_BASE_URL", "https://openrouter.ai/api")

	cfg, err := LoadConfig("INSPECTOR")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Provider != "openrouter" {
		t.Errorf("Provider = %q, want %q", cfg.Provider, "openrouter")
	}
	if cfg.Model != "qwen/qwen3.5-122b-a10b" {
		t.Errorf("Model = %q, want %q", cfg.Model, "qwen/qwen3.5-122b-a10b")
	}
	if cfg.APIKey != "sk-test" {
		t.Errorf("APIKey = %q, want %q", cfg.APIKey, "sk-test")
	}
	if cfg.BaseURL != "https://openrouter.ai/api" {
		t.Errorf("BaseURL = %q, want %q", cfg.BaseURL, "https://openrouter.ai/api")
	}
}

func TestLoadConfig_FallbackToFactory(t *testing.T) {
	clearEnv(t,
		"INSPECTOR_PROVIDER", "INSPECTOR_MODEL", "INSPECTOR_API_KEY", "INSPECTOR_BASE_URL",
		"FACTORY_PROVIDER", "FACTORY_MODEL", "FACTORY_API_KEY", "FACTORY_BASE_URL",
	)
	t.Setenv("FACTORY_PROVIDER", "openrouter")
	t.Setenv("FACTORY_MODEL", "qwen/qwen3.5-122b-a10b")
	t.Setenv("FACTORY_API_KEY", "sk-factory")

	cfg, err := LoadConfig("INSPECTOR")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Provider != "openrouter" {
		t.Errorf("Provider = %q, want %q", cfg.Provider, "openrouter")
	}
	if cfg.APIKey != "sk-factory" {
		t.Errorf("APIKey = %q, want %q", cfg.APIKey, "sk-factory")
	}
}

func TestLoadConfig_PrefixOverridesFactory(t *testing.T) {
	clearEnv(t,
		"INSPECTOR_PROVIDER", "INSPECTOR_MODEL", "INSPECTOR_API_KEY", "INSPECTOR_BASE_URL",
		"FACTORY_PROVIDER", "FACTORY_MODEL", "FACTORY_API_KEY", "FACTORY_BASE_URL",
	)
	t.Setenv("FACTORY_PROVIDER", "anthropic")
	t.Setenv("FACTORY_MODEL", "claude-sonnet-4-6")
	t.Setenv("INSPECTOR_PROVIDER", "openrouter")
	t.Setenv("INSPECTOR_MODEL", "qwen/qwen3.5-122b-a10b")

	cfg, err := LoadConfig("INSPECTOR")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Provider != "openrouter" {
		t.Errorf("Provider = %q, want %q", cfg.Provider, "openrouter")
	}
	if cfg.Model != "qwen/qwen3.5-122b-a10b" {
		t.Errorf("Model = %q, want %q", cfg.Model, "qwen/qwen3.5-122b-a10b")
	}
}

func TestLoadConfig_MissingProviderAndModel(t *testing.T) {
	clearEnv(t,
		"INSPECTOR_PROVIDER", "INSPECTOR_MODEL", "INSPECTOR_API_KEY", "INSPECTOR_BASE_URL",
		"FACTORY_PROVIDER", "FACTORY_MODEL", "FACTORY_API_KEY", "FACTORY_BASE_URL",
	)

	_, err := LoadConfig("INSPECTOR")
	if err == nil {
		t.Fatal("expected error for missing provider and model, got nil")
	}
}
