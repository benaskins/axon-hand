package hand

import (
	"fmt"
	"io"
	"math/rand/v2"
)

// Identity holds an agent's name, role, and version.
type Identity struct {
	Name    string
	Role    string
	Version string
}

// NewIdentity creates an Identity. If supplied is empty, a random
// adjective-noun name is generated.
func NewIdentity(role, version, supplied string) Identity {
	name := supplied
	if name == "" {
		name = generateName()
	}
	return Identity{Name: name, Role: role, Version: version}
}

// Banner writes the agent startup line to w.
func Banner(w io.Writer, id Identity) {
	fmt.Fprintf(w, "%s %s starting (worker: %s)\n", id.Role, id.Version, id.Name)
}

func generateName() string {
	adj := adjectives[rand.IntN(len(adjectives))]
	noun := nouns[rand.IntN(len(nouns))]
	return adj + "-" + noun
}

var adjectives = []string{
	"bold", "calm", "deft", "even", "fair",
	"glad", "hale", "keen", "live", "neat",
	"open", "pure", "rare", "safe", "true",
	"vast", "warm", "wise", "able", "deep",
	"fast", "good", "high", "just", "kind",
	"long", "mild", "nice", "pale", "rich",
	"slim", "soft", "tall", "wide", "wild",
	"dark", "firm", "full", "lean", "taut",
}

var nouns = []string{
	"arch", "beam", "bolt", "card", "chip",
	"cord", "dock", "drum", "edge", "fern",
	"fork", "gear", "grip", "harp", "helm",
	"iron", "jade", "knot", "lark", "leaf",
	"lens", "link", "loom", "mast", "mint",
	"node", "opal", "palm", "pine", "plum",
	"post", "reed", "ring", "root", "rust",
	"sail", "sand", "spar", "stem", "tide",
}
