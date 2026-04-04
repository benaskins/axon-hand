package hand

import (
	"time"

	"github.com/alecthomas/kong"
)

// CLI provides the common flags that every factory agent supports.
// Agents embed this struct and add their own fields.
type CLI struct {
	Name    string        `kong:"flag,help='Worker name (random adjective-noun if omitted)',short='n'"`
	Verbose bool          `kong:"flag,help='Verbose output to stderr',short='v'"`
	Timeout time.Duration `kong:"flag,default='15m',help='Operation timeout'"`
}

// ParseCLI parses command-line arguments into dest using kong. The dest
// struct should embed CLI for the common flags. The role and version are
// used for the app name and version in help output.
func ParseCLI(role, version string, dest any, args []string) error {
	parser, err := kong.New(dest,
		kong.Name(role),
		kong.Vars{"version": version},
	)
	if err != nil {
		return err
	}
	_, err = parser.Parse(args)
	return err
}
