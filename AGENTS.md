# AGENTS.md

## Module Selections

| Module | Purpose |
|---|---|
| axon-talk | LLM provider adapters (Anthropic, OpenAI-compatible) |
| kong | CLI argument parsing (struct-based, zero transitive deps) |

## Boundary Map

```
hand.Run (det)
  -> hand.ParseCLI (det) -> kong
  -> hand.LoadConfig (det) -> os.Getenv
  -> hand.NewClient (det) -> axon-talk/anthropic or axon-talk/openai
  -> hand.NewIdentity (det, randomness in naming)
  -> hand.Banner (det) -> io.Writer
  -> fn (non-det) -> agent-specific logic with LLM

hallmark.Emit (det)
  -> validate Input (det) -> closed enums, judge.model rule, value types
  -> newID (det, crypto/rand)
  -> otel.Tracer(...).Start (det) -> span attributes per the schema RFC
```

## Dependency Graph

```
axon-hand
  +-- axon-talk (LLM client interface + providers)
  +-- kong (CLI parsing)
  +-- hallmark (sub-package, OTEL span emission)
       +-- go.opentelemetry.io/otel (span API)
```

## Packages

- `hand` (root): chassis for CLI agents.
- `hallmark`: emitter for hallmark spans carrying a judge's
  observations about one artefact. Schema reference:
  werkhaus/docs/plans/2026-04-23-hallmark-span-schema.md. Producer
  identity is not on the hallmark; aggregators reach it via the
  artefact.
