package hand

import (
	"testing"
)

func TestNewClient_Anthropic(t *testing.T) {
	cfg := Config{Provider: "anthropic", Model: "claude-sonnet-4-6", APIKey: "sk-test"}
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClient_AnthropicDefaultURL(t *testing.T) {
	cfg := Config{Provider: "anthropic", Model: "claude-sonnet-4-6", APIKey: "sk-test"}
	_, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewClient_OpenRouter(t *testing.T) {
	cfg := Config{Provider: "openrouter", Model: "qwen/qwen3.5-122b-a10b", APIKey: "sk-or-test"}
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClient_OpenRouterCustomURL(t *testing.T) {
	cfg := Config{
		Provider: "openrouter",
		Model:    "qwen/qwen3.5-122b-a10b",
		APIKey:   "sk-or-test",
		BaseURL:  "https://custom.openrouter.example/api",
	}
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClient_Local(t *testing.T) {
	cfg := Config{Provider: "local", Model: "llama3.2"}
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClient_UnknownProvider(t *testing.T) {
	cfg := Config{Provider: "unknown", Model: "foo"}
	_, err := NewClient(cfg)
	if err == nil {
		t.Fatal("expected error for unknown provider, got nil")
	}
}
