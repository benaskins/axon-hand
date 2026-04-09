package hand

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	talk "github.com/benaskins/axon-talk"
)

// resultKey is the context key for the result accumulator.
type resultKey struct{}

// result accumulates agent output and token usage during execution.
// The chassis emits it as a JSON envelope on stdout after the agent returns.
type result struct {
	mu     sync.Mutex
	output string
	usage  *talk.Usage
}

// withResult attaches a result accumulator to the context.
func withResult(ctx context.Context) (context.Context, *result) {
	r := &result{}
	return context.WithValue(ctx, resultKey{}, r), r
}

// getResult retrieves the result accumulator from the context.
func getResult(ctx context.Context) *result {
	r, _ := ctx.Value(resultKey{}).(*result)
	return r
}

// SetOutput records the agent's primary output (e.g. a directory path,
// feedback text, or summary). The chassis includes it in the JSON
// envelope emitted on stdout. Call this instead of fmt.Println.
func SetOutput(ctx context.Context, output string) {
	if r := getResult(ctx); r != nil {
		r.mu.Lock()
		r.output = output
		r.mu.Unlock()
	}
}

// ReportUsage records token usage from the agent's LLM interactions.
// The chassis includes it in the JSON envelope emitted on stdout.
// Call this with the Usage from loop.Run or similar.
func ReportUsage(ctx context.Context, usage *talk.Usage) {
	if r := getResult(ctx); r != nil {
		r.mu.Lock()
		r.usage = usage
		r.mu.Unlock()
	}
}

// envelope is the JSON structure emitted on stdout by the chassis.
type envelope struct {
	Output     string     `json:"output,omitempty"`
	DurationMs int64      `json:"duration_ms"`
	Tokens     *tokenData `json:"tokens,omitempty"`
}

type tokenData struct {
	Input         int `json:"input"`
	Output        int `json:"output"`
	CacheCreation int `json:"cache_creation"`
	CacheRead     int `json:"cache_read"`
}

// writeEnvelope emits the JSON envelope to w.
func writeEnvelope(w io.Writer, durationMs int64, r *result) {
	env := envelope{DurationMs: durationMs}
	if r != nil {
		env.Output = r.output
		if r.usage != nil {
			env.Tokens = &tokenData{
				Input:         r.usage.InputTokens,
				Output:        r.usage.OutputTokens,
				CacheCreation: r.usage.CacheCreationInputTokens,
				CacheRead:     r.usage.CacheReadInputTokens,
			}
		}
	}
	data, err := json.Marshal(env)
	if err != nil {
		return
	}
	fmt.Fprintln(w, string(data))
}
