package hand

import (
	"bytes"
	"context"
	"testing"

	talk "github.com/benaskins/axon-talk"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestRunWith_DisableTraceLeavesClientUnwrapped(t *testing.T) {
	clearEnv(t,
		"TEST_PROVIDER", "TEST_MODEL", "TEST_API_KEY", "TEST_BASE_URL",
		"FACTORY_PROVIDER", "FACTORY_MODEL", "FACTORY_API_KEY", "FACTORY_BASE_URL",
	)
	t.Setenv("FACTORY_PROVIDER", "anthropic")
	t.Setenv("FACTORY_MODEL", "m")
	t.Setenv("FACTORY_API_KEY", "sk-test")

	var got talk.LLMClient
	var stderr, stdout bytes.Buffer

	code := RunWith(RunConfig{
		Role: "test", Version: "0.1.0",
		Args: []string{"--name", "keen-walnut"}, Stderr: &stderr, Stdout: &stdout,
		DisableTrace: true,
		Fn: func(ctx context.Context, id Identity, client talk.LLMClient) error {
			got = client
			return nil
		},
	})
	if code != 0 {
		t.Fatalf("exit=%d stderr=%s", code, stderr.String())
	}
	if got == nil {
		t.Fatal("client nil")
	}
}

func TestRunWith_EmitsAgentSpan(t *testing.T) {
	clearEnv(t,
		"TEST_PROVIDER", "TEST_MODEL", "TEST_API_KEY", "TEST_BASE_URL",
		"FACTORY_PROVIDER", "FACTORY_MODEL", "FACTORY_API_KEY", "FACTORY_BASE_URL",
	)
	t.Setenv("FACTORY_PROVIDER", "anthropic")
	t.Setenv("FACTORY_MODEL", "m")
	t.Setenv("FACTORY_API_KEY", "sk-test")

	// Install a recorder BEFORE calling RunWith. With DisableTrace=true the
	// chassis leaves our provider untouched.
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	otel.SetTracerProvider(tp)
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	var stderr, stdout bytes.Buffer
	code := RunWith(RunConfig{
		Role: "plan-lead", Version: "0.1.0",
		Args: []string{"--name", "quiet-oak"}, Stderr: &stderr, Stdout: &stdout,
		DisableTrace: true, // keep our recorder active
		Fn: func(ctx context.Context, id Identity, client talk.LLMClient) error {
			return nil
		},
	})
	if code != 0 {
		t.Fatalf("exit=%d stderr=%s", code, stderr.String())
	}

	spans := recorder.Ended()
	if len(spans) != 1 {
		t.Fatalf("spans=%d, want 1", len(spans))
	}
	if spans[0].Name() != "agent.plan-lead" {
		t.Errorf("span name = %q, want agent.plan-lead", spans[0].Name())
	}
}
