# axon-agent -- Initial Build Plan
# 2026-04-04

Each step is commit-sized. Execute via `/iterate`.

## Step 1 -- Scaffold project and config loading

Create the module skeleton: go.mod, justfile (build/test/vet), README.md
stub. Implement `config.go` with `Config` struct (Provider, Model,
APIKey, BaseURL) and `LoadConfig(prefix string) (Config, error)`. Reads
`{PREFIX}_PROVIDER`, `{PREFIX}_MODEL`, `{PREFIX}_API_KEY`,
`{PREFIX}_BASE_URL` from env, falling back to `FACTORY_PROVIDER`,
`FACTORY_MODEL`, `FACTORY_API_KEY`, `FACTORY_BASE_URL`. Returns error
if Provider and Model are both empty after fallback. Write table-driven
tests covering: prefix vars set, fallback to FACTORY vars, prefix
overrides factory, missing provider/model errors.

Commit: `feat: scaffold axon-agent with env-based config loading and FACTORY_* fallback`

## Step 2 -- LLM client construction

Implement `NewClient(cfg Config) (talk.LLMClient, error)`. Switch on
Provider: "anthropic" uses axon-talk/anthropic with BaseURL defaulting
to https://api.anthropic.com, "openrouter" uses axon-talk/openai with
BaseURL defaulting to https://openrouter.ai/api, "local" uses
axon-talk/openai with BaseURL defaulting to http://localhost:11434.
Unknown provider returns error. Write tests that verify: correct client
type returned for each provider, default URLs applied, custom BaseURL
honoured, unknown provider errors.

Commit: `feat: add NewClient with anthropic, openrouter, and local provider support`

## Step 3 -- Agent identity and naming

Implement `identity.go` with `Identity` struct (Name, Role, Version)
and `NewIdentity(role, version, supplied string) Identity`. If supplied
name is empty, generate a random adjective-noun pair from embedded word
lists. Implement `Banner(w io.Writer, id Identity)` that prints the
startup line to the given writer (e.g. "inspector 0.1.0 starting
(worker: keen-walnut)"). Write tests: supplied name returned as-is,
generated name matches adjective-noun pattern, banner format correct.

Commit: `feat: add Identity with adjective-noun naming and startup banner`

## Step 4 -- Kong CLI base struct

Implement `cli.go` with a `CLI` struct using kong tags: Name (string,
optional flag), Verbose (bool flag), Timeout (duration flag, default
15m), ProjectDir (positional arg, optional since not all agents need
it). Implement `ParseCLI(role, version string, dest any) error` that
creates a kong parser with the agent role as the app name and parses
os.Args. The dest struct should embed CLI for the common fields. Write
tests: parse with all flags set, parse with defaults, parse with
positional arg.

Commit: `feat: add kong-based CLI struct with common agent flags`

## Step 5 -- Lifecycle runner

Implement `run.go` with `Run(role, version string, cli any, fn
func(context.Context, Identity, talk.LLMClient) error)`. Run handles:
parse CLI (via ParseCLI into cli), load config (using role as prefix),
build client (via NewClient), construct identity (from cli.Name, role,
version), print banner to stderr, create context with signal handling
(SIGINT/SIGTERM), call fn, and exit with code 0/1/2 based on result.
Config errors exit 2, fn errors exit 1. Write tests using a mock fn
that verifies identity and client are passed through. Test signal
cancellation with a goroutine.

Commit: `feat: add Run lifecycle with config, client, identity, and signal handling`

## Step 6 -- Documentation

Write README.md with usage examples showing how an agent imports and
uses the chassis. Show the minimal agent (embed CLI, call Run, use
client in fn). Document env var convention. Write AGENTS.md with module
selections and boundary map.

Commit: `docs: write README and AGENTS.md`
