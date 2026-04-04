package hand

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	talk "github.com/benaskins/axon-talk"
)

// AgentFunc is the function signature for an agent's main logic.
type AgentFunc func(ctx context.Context, id Identity, client talk.LLMClient) error

// RunConfig configures RunWith. Tests use this to inject args and stderr.
type RunConfig struct {
	Role    string
	Version string
	Args    []string
	Stderr  io.Writer
	Fn      AgentFunc
}

// RunWith executes the agent lifecycle and returns the exit code.
// Intended for testing; production agents use Run.
func RunWith(rc RunConfig) int {
	stderr := rc.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	// Parse CLI.
	var cli struct {
		CLI
	}
	if err := ParseCLI(rc.Role, rc.Version, &cli, rc.Args); err != nil {
		fmt.Fprintf(stderr, "%s: %v\n", rc.Role, err)
		return 2
	}

	// Load config.
	cfg, err := LoadConfig(strings.ToUpper(rc.Role))
	if err != nil {
		fmt.Fprintf(stderr, "%s: %v\n", rc.Role, err)
		return 2
	}

	// Build client.
	client, err := NewClient(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "%s: %v\n", rc.Role, err)
		return 2
	}

	// Identity and banner.
	id := NewIdentity(rc.Role, rc.Version, cli.Name)
	Banner(stderr, id)

	// Context with signal handling.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Run the agent.
	if err := rc.Fn(ctx, id, client); err != nil {
		fmt.Fprintf(stderr, "%s: error: %v\n", rc.Role, err)
		return 1
	}

	return 0
}

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
