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
```

## Dependency Graph

```
axon-hand
  +-- axon-talk (LLM client interface + providers)
  +-- kong (CLI parsing)
```
