package hand

import (
	"fmt"
	"strings"

	talk "github.com/benaskins/axon-talk"
	"github.com/benaskins/axon-talk/anthropic"
	"github.com/benaskins/axon-talk/openai"
)

// NewClient constructs a talk.LLMClient from the given Config.
//
// Supported providers:
//   - "anthropic": Anthropic API (default BaseURL: https://api.anthropic.com)
//   - "openrouter": OpenRouter (default BaseURL: https://openrouter.ai/api)
//   - "local": OpenAI-compatible local server (default BaseURL: http://localhost:11434)
func NewClient(cfg Config) (talk.LLMClient, error) {
	switch strings.ToLower(cfg.Provider) {
	case "anthropic":
		baseURL := cfg.BaseURL
		if baseURL == "" {
			baseURL = "https://api.anthropic.com"
		}
		return anthropic.NewClient(baseURL, cfg.APIKey), nil

	case "openrouter":
		baseURL := cfg.BaseURL
		if baseURL == "" {
			baseURL = "https://openrouter.ai/api"
		}
		return openai.NewClient(baseURL, cfg.APIKey), nil

	case "local":
		baseURL := cfg.BaseURL
		if baseURL == "" {
			baseURL = "http://localhost:11434"
		}
		return openai.NewClient(baseURL, cfg.APIKey), nil

	default:
		return nil, fmt.Errorf("hand: unsupported provider %q (expected anthropic, openrouter, or local)", cfg.Provider)
	}
}
