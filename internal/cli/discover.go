package cli

import (
	"fmt"
	"path/filepath"

	"github.com/vyper-tooling/pi/internal/builtins"
	"github.com/vyper-tooling/pi/internal/discovery"
)

// discoverAll discovers local automations from .pi/ and merges built-in
// automations. Local automations take precedence over built-ins with the
// same name.
func discoverAll(root string) (*discovery.Result, error) {
	piDir := filepath.Join(root, discovery.PiDir)
	result, err := discovery.Discover(piDir)
	if err != nil {
		return nil, fmt.Errorf("discovering automations: %w", err)
	}

	builtinResult, err := builtins.Discover()
	if err != nil {
		return nil, fmt.Errorf("discovering built-in automations: %w", err)
	}

	result.MergeBuiltins(builtinResult)

	return result, nil
}
