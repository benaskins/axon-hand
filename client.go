package hand

import (
	"fmt"
	"os"
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

// NewClientWithIdentity constructs a talk.LLMClient with identity and
// telemetry headers for request tracing and cost attribution.
//
// Headers sent on every request:
//   - X-Title: role/instance (e.g. "mech-hand/bold-elm")
//   - HTTP-Referer: werkhaus:card or werkhaus:role
//   - X-Werk-Card: board card ID (from WERK_CARD env)
//   - X-Werk-Instance: worker instance name (from WERK_INSTANCE env)
//   - X-Werk-Attempt: retry attempt number (from WERK_ATTEMPT env)
//   - X-Werk-Pipeline: pipeline state (from WERK_STATE env)
func NewClientWithIdentity(cfg Config, id Identity) (talk.LLMClient, error) {
	headers := make(map[string]string)

	// Identity
	if id.Role != "" {
		headers["X-Title"] = id.Role + "/" + id.Name
	}

	// Telemetry from pipeline env (set by werk run)
	card := os.Getenv("WERK_CARD")
	if card != "" {
		headers["HTTP-Referer"] = "werkhaus:" + card
		headers["X-Werk-Card"] = card
	} else if id.Role != "" {
		headers["HTTP-Referer"] = "werkhaus:" + id.Role
	}

	if v := os.Getenv("WERK_INSTANCE"); v != "" {
		headers["X-Werk-Instance"] = v
	} else if id.Name != "" {
		headers["X-Werk-Instance"] = id.Name
	}

	if v := os.Getenv("WERK_ATTEMPT"); v != "" {
		headers["X-Werk-Attempt"] = v
	}
	if v := os.Getenv("WERK_STATE"); v != "" {
		headers["X-Werk-Pipeline"] = v
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
		if len(headers) > 0 {
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
