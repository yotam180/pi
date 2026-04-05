package builtins

import (
	"fmt"

	"github.com/vyper-tooling/pi/internal/automation"
	pisemver "github.com/vyper-tooling/pi/internal/semver"
)

// newVersionSatisfies creates the pi:version-satisfies Go-backed builtin.
// It takes two inputs: "version" (the actual version string) and "required"
// (the constraint expression). Returns exit code 0 if satisfied, 1 if not.
func newVersionSatisfies() *automation.Automation {
	a := &automation.Automation{
		Name:        "version-satisfies",
		Description: "Check whether a version string satisfies a constraint (semver-aware)",
		Inputs: map[string]automation.InputSpec{
			"version": {
				Type:        "string",
				Description: "Version string to check (e.g. \"22.3.1\")",
			},
			"required": {
				Type:        "string",
				Description: "Constraint expression (e.g. \"22\", \">=20\", \"^18\", \"~20.1\")",
			},
		},
		InputKeys: []string{"version", "required"},
		GoFunc: func(inputs map[string]string) error {
			version := inputs["version"]
			required := inputs["required"]

			if version == "" {
				return fmt.Errorf("version-satisfies: version input is empty")
			}
			if required == "" {
				return fmt.Errorf("version-satisfies: required input is empty")
			}

			if err := pisemver.Satisfies(version, required); err != nil {
				return &goFuncExitError{msg: err.Error()}
			}
			return nil
		},
	}
	return a
}

// goFuncExitError signals a non-zero exit from a Go-backed builtin.
// The executor treats this as a step failure (exit code 1).
type goFuncExitError struct {
	msg string
}

func (e *goFuncExitError) Error() string {
	return e.msg
}
