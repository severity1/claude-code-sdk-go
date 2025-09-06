package claudecode

import (
	"github.com/severity1/claude-code-sdk-go/internal/cli"
	"github.com/severity1/claude-code-sdk-go/internal/shared"
	"github.com/severity1/claude-code-sdk-go/internal/subprocess"
)

// init sets up the default transport factory to create subprocess transports.
func init() {
	defaultTransportFactory = func(options *Options, closeStdin bool) (Transport, error) {
		// Find Claude CLI binary
		cliPath, err := cli.FindCLI()
		if err != nil {
			return nil, err
		}

		// Create subprocess transport
		return subprocess.New(cliPath, (*shared.Options)(options), closeStdin), nil
	}
}
