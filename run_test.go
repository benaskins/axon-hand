package hand

import (
	"bytes"
	"context"
	"errors"
	"testing"

	talk "github.com/benaskins/axon-talk"
)

func TestRun_PassesThroughIdentityAndClient(t *testing.T) {
	clearEnv(t,
		"TEST_PROVIDER", "TEST_MODEL", "TEST_API_KEY", "TEST_BASE_URL",
		"FACTORY_PROVIDER", "FACTORY_MODEL", "FACTORY_API_KEY", "FACTORY_BASE_URL",
	)
	t.Setenv("FACTORY_PROVIDER", "anthropic")
	t.Setenv("FACTORY_MODEL", "claude-sonnet-4-6")
	t.Setenv("FACTORY_API_KEY", "sk-test")

	var gotID Identity
	var gotClient talk.LLMClient
	var stderr bytes.Buffer

	code := RunWith(RunConfig{
		Role:    "test",
		Version: "0.1.0",
		Args:    []string{"--name", "keen-walnut"},
		Stderr:  &stderr,
		Fn: func(ctx context.Context, id Identity, client talk.LLMClient) error {
			gotID = id
			gotClient = client
			return nil
		},
	})

	if code != 0 {
		t.Errorf("exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	if gotID.Name != "keen-walnut" {
		t.Errorf("Identity.Name = %q, want %q", gotID.Name, "keen-walnut")
	}
	if gotID.Role != "test" {
		t.Errorf("Identity.Role = %q, want %q", gotID.Role, "test")
	}
	if gotClient == nil {
		t.Error("expected non-nil client")
	}
	if !bytes.Contains(stderr.Bytes(), []byte("keen-walnut")) {
		t.Errorf("stderr missing worker name: %s", stderr.String())
	}
}

func TestRun_ConfigError(t *testing.T) {
	clearEnv(t,
		"TEST_PROVIDER", "TEST_MODEL", "TEST_API_KEY", "TEST_BASE_URL",
		"FACTORY_PROVIDER", "FACTORY_MODEL", "FACTORY_API_KEY", "FACTORY_BASE_URL",
	)

	var stderr bytes.Buffer
	code := RunWith(RunConfig{
		Role:    "test",
		Version: "0.1.0",
		Args:    []string{},
		Stderr:  &stderr,
		Fn: func(ctx context.Context, id Identity, client talk.LLMClient) error {
			t.Fatal("fn should not be called on config error")
			return nil
		},
	})

	if code != 2 {
		t.Errorf("exit code = %d, want 2", code)
	}
}

func TestRun_FnError(t *testing.T) {
	clearEnv(t,
		"TEST_PROVIDER", "TEST_MODEL", "TEST_API_KEY", "TEST_BASE_URL",
		"FACTORY_PROVIDER", "FACTORY_MODEL", "FACTORY_API_KEY", "FACTORY_BASE_URL",
	)
	t.Setenv("FACTORY_PROVIDER", "anthropic")
	t.Setenv("FACTORY_MODEL", "test-model")
	t.Setenv("FACTORY_API_KEY", "sk-test")

	var stderr bytes.Buffer
	code := RunWith(RunConfig{
		Role:    "test",
		Version: "0.1.0",
		Args:    []string{},
		Stderr:  &stderr,
		Fn: func(ctx context.Context, id Identity, client talk.LLMClient) error {
			return errors.New("something went wrong")
		},
	})

	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
}

func TestRun_ContextCancellation(t *testing.T) {
	clearEnv(t,
		"TEST_PROVIDER", "TEST_MODEL", "TEST_API_KEY", "TEST_BASE_URL",
		"FACTORY_PROVIDER", "FACTORY_MODEL", "FACTORY_API_KEY", "FACTORY_BASE_URL",
	)
	t.Setenv("FACTORY_PROVIDER", "anthropic")
	t.Setenv("FACTORY_MODEL", "test-model")
	t.Setenv("FACTORY_API_KEY", "sk-test")

	var stderr bytes.Buffer
	code := RunWith(RunConfig{
		Role:    "test",
		Version: "0.1.0",
		Args:    []string{},
		Stderr:  &stderr,
		Fn: func(ctx context.Context, id Identity, client talk.LLMClient) error {
			if ctx == nil {
				t.Error("expected non-nil context")
			}
			return nil
		},
	})

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
}
