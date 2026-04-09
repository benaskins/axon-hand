package hand

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	talk "github.com/benaskins/axon-talk"
)

func TestEnvelope_EmittedOnSuccess(t *testing.T) {
	clearEnv(t,
		"TEST_PROVIDER", "TEST_MODEL", "TEST_API_KEY", "TEST_BASE_URL",
		"FACTORY_PROVIDER", "FACTORY_MODEL", "FACTORY_API_KEY", "FACTORY_BASE_URL",
	)
	t.Setenv("FACTORY_PROVIDER", "anthropic")
	t.Setenv("FACTORY_MODEL", "test-model")
	t.Setenv("FACTORY_API_KEY", "sk-test")

	var stderr, stdout bytes.Buffer
	code := RunWith(RunConfig{
		Role:    "test",
		Version: "0.1.0",
		Stderr:  &stderr,
		Stdout:  &stdout,
		Fn: func(ctx context.Context, id Identity, client talk.LLMClient) error {
			SetOutput(ctx, "/some/path")
			ReportUsage(ctx, &talk.Usage{
				InputTokens:  1000,
				OutputTokens: 200,
			})
			return nil
		},
	})

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr.String())
	}

	var env envelope
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("failed to parse envelope: %v\nraw: %s", err, stdout.String())
	}

	if env.Output != "/some/path" {
		t.Errorf("Output = %q, want %q", env.Output, "/some/path")
	}
	if env.DurationMs < 0 {
		t.Errorf("DurationMs = %d, want >= 0", env.DurationMs)
	}
	if env.Tokens == nil {
		t.Fatal("Tokens is nil, want non-nil")
	}
	if env.Tokens.Input != 1000 {
		t.Errorf("Tokens.Input = %d, want 1000", env.Tokens.Input)
	}
	if env.Tokens.Output != 200 {
		t.Errorf("Tokens.Output = %d, want 200", env.Tokens.Output)
	}
}

func TestEnvelope_EmittedOnError(t *testing.T) {
	clearEnv(t,
		"TEST_PROVIDER", "TEST_MODEL", "TEST_API_KEY", "TEST_BASE_URL",
		"FACTORY_PROVIDER", "FACTORY_MODEL", "FACTORY_API_KEY", "FACTORY_BASE_URL",
	)
	t.Setenv("FACTORY_PROVIDER", "anthropic")
	t.Setenv("FACTORY_MODEL", "test-model")
	t.Setenv("FACTORY_API_KEY", "sk-test")

	var stderr, stdout bytes.Buffer
	code := RunWith(RunConfig{
		Role:    "test",
		Version: "0.1.0",
		Stderr:  &stderr,
		Stdout:  &stdout,
		Fn: func(ctx context.Context, id Identity, client talk.LLMClient) error {
			ReportUsage(ctx, &talk.Usage{InputTokens: 500})
			return context.Canceled
		},
	})

	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}

	var env envelope
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("failed to parse envelope: %v\nraw: %s", err, stdout.String())
	}

	if env.Tokens == nil || env.Tokens.Input != 500 {
		t.Errorf("expected tokens.input=500 even on error, got %+v", env.Tokens)
	}
}

func TestEnvelope_NoUsageOmitsTokens(t *testing.T) {
	clearEnv(t,
		"TEST_PROVIDER", "TEST_MODEL", "TEST_API_KEY", "TEST_BASE_URL",
		"FACTORY_PROVIDER", "FACTORY_MODEL", "FACTORY_API_KEY", "FACTORY_BASE_URL",
	)
	t.Setenv("FACTORY_PROVIDER", "anthropic")
	t.Setenv("FACTORY_MODEL", "test-model")
	t.Setenv("FACTORY_API_KEY", "sk-test")

	var stderr, stdout bytes.Buffer
	code := RunWith(RunConfig{
		Role:    "test",
		Version: "0.1.0",
		Stderr:  &stderr,
		Stdout:  &stdout,
		Fn: func(ctx context.Context, id Identity, client talk.LLMClient) error {
			SetOutput(ctx, "done")
			return nil
		},
	})

	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}

	var env envelope
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("failed to parse envelope: %v\nraw: %s", err, stdout.String())
	}

	if env.Output != "done" {
		t.Errorf("Output = %q, want %q", env.Output, "done")
	}
	if env.Tokens != nil {
		t.Errorf("Tokens = %+v, want nil (omitted)", env.Tokens)
	}
}

func TestEnvelope_CacheTokens(t *testing.T) {
	clearEnv(t,
		"TEST_PROVIDER", "TEST_MODEL", "TEST_API_KEY", "TEST_BASE_URL",
		"FACTORY_PROVIDER", "FACTORY_MODEL", "FACTORY_API_KEY", "FACTORY_BASE_URL",
	)
	t.Setenv("FACTORY_PROVIDER", "anthropic")
	t.Setenv("FACTORY_MODEL", "test-model")
	t.Setenv("FACTORY_API_KEY", "sk-test")

	var stderr, stdout bytes.Buffer
	code := RunWith(RunConfig{
		Role:    "test",
		Version: "0.1.0",
		Stderr:  &stderr,
		Stdout:  &stdout,
		Fn: func(ctx context.Context, id Identity, client talk.LLMClient) error {
			ReportUsage(ctx, &talk.Usage{
				InputTokens:              1000,
				OutputTokens:             200,
				CacheCreationInputTokens: 5000,
				CacheReadInputTokens:     3000,
			})
			return nil
		},
	})

	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}

	var env envelope
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("failed to parse envelope: %v", err)
	}

	if env.Tokens.CacheCreation != 5000 {
		t.Errorf("CacheCreation = %d, want 5000", env.Tokens.CacheCreation)
	}
	if env.Tokens.CacheRead != 3000 {
		t.Errorf("CacheRead = %d, want 3000", env.Tokens.CacheRead)
	}
}
