package hand

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	looktrace "github.com/benaskins/axon-look/trace"
	talk "github.com/benaskins/axon-talk"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// AgentFunc is the function signature for an agent's main logic.
type AgentFunc func(ctx context.Context, id Identity, client talk.LLMClient) error

// RunConfig configures RunWith. Tests use this to inject args and stderr.
type RunConfig struct {
	Role    string
	Version string
	Args    []string
	Stderr  io.Writer
	Stdout  io.Writer // Where the JSON envelope is written. Defaults to os.Stdout.
	CLI     any       // Agent's CLI struct (must embed hand.CLI). If nil, a default is used.
	Fn      AgentFunc

	// DisableTrace skips OTEL tracer installation and LLMClient wrapping.
	// Tests use this to keep assertions clean; production agents leave it
	// false so every chassis run emits an agent.<role> span with llm.call
	// children.
	DisableTrace bool
}

// RunWith executes the agent lifecycle and returns the exit code.
// Intended for testing; production agents use Run.
func RunWith(rc RunConfig) int {
	stderr := rc.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}
	stdout := rc.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}

	// Use provided CLI struct or default.
	cliDest := rc.CLI
	if cliDest == nil {
		cliDest = &struct{ CLI }{}
	}
	if err := ParseCLI(rc.Role, rc.Version, cliDest, rc.Args); err != nil {
		fmt.Fprintf(stderr, "%s: %v\n", rc.Role, err)
		return 2
	}

	// Extract the embedded CLI base.
	base := extractCLI(cliDest)

	// Load config.
	cfg, err := LoadConfig(strings.ToUpper(rc.Role))
	if err != nil {
		fmt.Fprintf(stderr, "%s: %v\n", rc.Role, err)
		return 2
	}

	// Identity and banner.
	id := NewIdentity(rc.Role, rc.Version, base.Name)

	// Build client with identity headers for tracing.
	client, err := NewClientWithIdentity(cfg, id)
	if err != nil {
		fmt.Fprintf(stderr, "%s: %v\n", rc.Role, err)
		return 2
	}
	Banner(stderr, id)

	// Context with signal handling and result accumulator.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	ctx, res := withResult(ctx)

	// OTEL: install a tracer provider and wrap the client so every Chat
	// call emits an llm.call span. Open an agent.<role> span around Fn
	// so child llm.call spans inherit the agent context.
	if !rc.DisableTrace {
		shutdown, traceErr := looktrace.Init(ctx, rc.Role)
		if traceErr != nil {
			fmt.Fprintf(stderr, "%s: trace init: %v\n", rc.Role, traceErr)
		} else {
			defer func() { _ = shutdown(context.Background()) }()
			client = looktrace.WrapLLMClient(client)
		}
	}

	tracer := otel.Tracer("github.com/benaskins/axon-hand")
	ctx, span := tracer.Start(ctx, "agent."+rc.Role, withAgentAttrs(id))

	start := time.Now()

	// Run the agent.
	if err := rc.Fn(ctx, id, client); err != nil {
		durationMs := time.Since(start).Milliseconds()
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		span.End()
		writeEnvelope(stdout, durationMs, res)
		fmt.Fprintf(stderr, "%s: error: %v\n", rc.Role, err)
		return 1
	}

	durationMs := time.Since(start).Milliseconds()
	span.End()
	writeEnvelope(stdout, durationMs, res)
	return 0
}

func withAgentAttrs(id Identity) oteltrace.SpanStartOption {
	return oteltrace.WithAttributes(
		attribute.String("agent.role", id.Role),
		attribute.String("agent.name", id.Name),
		attribute.String("agent.version", id.Version),
	)
}

// extractCLI gets the embedded CLI from an agent's CLI struct.
func extractCLI(dest any) CLI {
	type hasCLI interface{ GetCLI() CLI }
	if h, ok := dest.(hasCLI); ok {
		return h.GetCLI()
	}
	// Try direct type assertion for the default struct.
	type defaultCLI struct{ CLI }
	if d, ok := dest.(*defaultCLI); ok {
		return d.CLI
	}
	return CLI{}
}

// GetCLI returns the CLI base. Agents whose CLI struct embeds hand.CLI
// automatically satisfy this via Go embedding.
func (c CLI) GetCLI() CLI { return c }

// Run is the production entry point. It parses CLI args from os.Args,
// loads config, builds the client, and calls fn. Exits with 0 on
// success, 1 on agent error, 2 on config error.
func Run(role, version string, fn AgentFunc) {
	code := RunWith(RunConfig{
		Role:    role,
		Version: version,
		Args:    os.Args[1:],
		Stderr:  os.Stderr,
		Fn:      fn,
	})
	os.Exit(code)
}

// RunCLI is like Run but accepts the agent's CLI struct for extended flags.
// The CLI struct must embed hand.CLI and be a pointer.
func RunCLI(role, version string, cli any, fn AgentFunc) {
	code := RunWith(RunConfig{
		Role:    role,
		Version: version,
		Args:    os.Args[1:],
		Stderr:  os.Stderr,
		CLI:     cli,
		Fn:      fn,
	})
	os.Exit(code)
}
