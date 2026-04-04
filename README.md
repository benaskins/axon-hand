# axon-hand

Shared chassis for factory floor agents. Provides LLM client
configuration, worker identity, CLI parsing (kong), and lifecycle
management.

## Usage

A minimal agent:

```go
package main

import (
    "context"
    "fmt"

    hand "github.com/benaskins/axon-hand"
    talk "github.com/benaskins/axon-talk"
)

func main() {
    hand.Run("myagent", "0.1.0", func(ctx context.Context, id hand.Identity, client talk.LLMClient) error {
        fmt.Fprintf(os.Stderr, "hello from %s\n", id.Name)
        // Use client with axon-loop/axon-tool here
        return nil
    })
}
```

Agents can extend the CLI with their own flags:

```go
var cli struct {
    hand.CLI
    ProjectDir string `kong:"arg,required,help='Project directory'"`
    Layers     string `kong:"flag,default='static,security,test',help='Layers to run'"`
}
```

## Environment Variables

Agent-specific prefix takes precedence, falling back to fleet-wide
`FACTORY_*` defaults.

| Variable | Description |
|---|---|
| `{PREFIX}_PROVIDER` / `FACTORY_PROVIDER` | Provider: anthropic, openrouter, local |
| `{PREFIX}_MODEL` / `FACTORY_MODEL` | Model name |
| `{PREFIX}_API_KEY` / `FACTORY_API_KEY` | API key |
| `{PREFIX}_BASE_URL` / `FACTORY_BASE_URL` | Base URL override |

## Exit Codes

| Code | Meaning |
|---|---|
| 0 | Success |
| 1 | Agent error (work failed) |
| 2 | Configuration error |

## Common Flags

| Flag | Default | Description |
|---|---|---|
| `--name` | random adjective-noun | Worker name |
| `--verbose` | false | Verbose output to stderr |
| `--timeout` | 15m | Operation timeout |
