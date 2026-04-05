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
	return NewClientWithIdentity(cfg, Identity{})
}

// NewClientWithIdentity constructs a talk.LLMClient with identity headers
// for request tracing and attribution.
func NewClientWithIdentity(cfg Config, id Identity) (talk.LLMClient, error) {
	// Build identity headers for OpenAI-compatible providers
	var headers map[string]string
	if id.Role != "" {
		headers = map[string]string{
			"X-Title":      id.Role + "/" + id.Name,
			"HTTP-Referer": "werkhaus:" + id.Role,
		}
	}

	switch strings.ToLower(cfg.Provider) {
	case "anthropic":
		baseURL := cfg.BaseURL
		if baseURL == "" {
			baseURL = "https://api.anthropic.com"
		}
		// Anthropic client uses its own header scheme; identity via metadata later
		return anthropic.NewClient(baseURL, cfg.APIKey), nil

	case "openrouter":
		baseURL := cfg.BaseURL
		if baseURL == "" {
			baseURL = "https://openrouter.ai/api"
		}
		var opts []openai.Option
		if headers != nil {
			opts = append(opts, openai.WithHeaders(headers))
		}
		return openai.NewClient(baseURL, cfg.APIKey, opts...), nil

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
