// Package agent provides a shared chassis for factory floor agents.
//
// Every factory agent imports this package for LLM client configuration,
// identity, CLI parsing, and lifecycle management.
package agent

import (
	"fmt"
	"os"
	"strings"
)

// Config holds LLM provider configuration for an agent.
type Config struct {
	Provider string
	Model    string
	APIKey   string
	BaseURL  string
}

// LoadConfig reads LLM configuration from environment variables. It checks
// {prefix}_PROVIDER, {prefix}_MODEL, {prefix}_API_KEY, {prefix}_BASE_URL
// first, falling back to FACTORY_PROVIDER, FACTORY_MODEL, FACTORY_API_KEY,
// FACTORY_BASE_URL for any that are unset.
//
// Returns an error if both Provider and Model are empty after fallback.
func LoadConfig(prefix string) (Config, error) {
	prefix = strings.ToUpper(prefix)

	cfg := Config{
		Provider: envWithFallback(prefix+"_PROVIDER", "FACTORY_PROVIDER"),
		Model:    envWithFallback(prefix+"_MODEL", "FACTORY_MODEL"),
		APIKey:   envWithFallback(prefix+"_API_KEY", "FACTORY_API_KEY"),
		BaseURL:  envWithFallback(prefix+"_BASE_URL", "FACTORY_BASE_URL"),
	}

	if cfg.Provider == "" && cfg.Model == "" {
		return Config{}, fmt.Errorf("agent config: %s_PROVIDER and %s_MODEL are both unset (FACTORY_* fallback also empty)", prefix, prefix)
	}

	return cfg, nil
}

func envWithFallback(primary, fallback string) string {
	if v := os.Getenv(primary); v != "" {
		return v
	}
	return os.Getenv(fallback)
}
