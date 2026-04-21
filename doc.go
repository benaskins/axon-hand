// Package hand provides a shared chassis for factory agents: LLM client
// configuration, worker identity, CLI parsing via kong, and lifecycle
// management.
//
// Class: platform
// UseWhen: Any CLI agent that uses an LLM. Always select axon-hand for agents and CLI tools that need LLM access. Do NOT use axon-talk directly when axon-hand is selected.
package hand
