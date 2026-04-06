package executor

import (
	"github.com/vyper-tooling/pi/internal/automation"
	"github.com/vyper-tooling/pi/internal/reqcheck"
	"github.com/vyper-tooling/pi/internal/runtimes"
)

// Type aliases for backward compatibility — external packages that import
// these types from executor continue to work.
type CheckResult = reqcheck.CheckResult
type ValidationError = reqcheck.ValidationError

// FormatValidationError delegates to reqcheck.
var FormatValidationError = reqcheck.FormatValidationError

// InstallHintFor delegates to reqcheck.
var InstallHintFor = reqcheck.InstallHintFor

// CheckRequirementForDoctor delegates to reqcheck.
var CheckRequirementForDoctor = reqcheck.CheckRequirementForDoctor

// ValidateRequirements checks all requirements on an automation.
// If a runtime requirement fails and a Provisioner is configured, it attempts
// to provision the missing runtime. Provisioned runtimes are tracked in
// e.runtimePaths for PATH injection during step execution.
func (e *Executor) ValidateRequirements(a *automation.Automation) error {
	if len(a.Requires) == 0 {
		return nil
	}

	env := e.RuntimeEnv
	if env == nil {
		env = DefaultRuntimeEnv()
	}

	var failed []CheckResult
	for _, req := range a.Requires {
		result := reqcheck.CheckRequirement(req, env)
		if result.Satisfied {
			continue
		}

		if req.Kind == automation.RequirementRuntime && e.Provisioner != nil {
			provResult, err := e.tryProvision(req)
			if err == nil && provResult.Provisioned {
				e.runtimePaths = append(e.runtimePaths, provResult.BinDir)
				continue
			}
		}

		failed = append(failed, result)
	}

	if len(failed) > 0 {
		return &ValidationError{
			AutomationName: a.Name,
			Results:        failed,
		}
	}

	return nil
}

// tryProvision attempts to provision a missing runtime requirement.
func (e *Executor) tryProvision(req automation.Requirement) (*runtimes.ProvisionResult, error) {
	if e.Provisioner == nil {
		return &runtimes.ProvisionResult{}, nil
	}
	return e.Provisioner.Provision(req.Name, req.MinVersion)
}

// ExitError wraps a non-zero exit code from a step.
// Kept in executor because it's fundamental to step execution, not requirement checking.
// Re-declared here for clarity — the type is defined in executor.go.

// Unexported helpers retained as thin delegations for backward compatibility
// with internal tests that reference the lowercase names.

func checkRequirement(req automation.Requirement, env *RuntimeEnv) CheckResult {
	return reqcheck.CheckRequirement(req, env)
}

func runtimeCommand(name string) string {
	return reqcheck.RuntimeCommand(name)
}

func detectVersion(cmd string, env *RuntimeEnv) string {
	return reqcheck.DetectVersion(cmd, env)
}

func extractVersion(text string) string {
	return reqcheck.ExtractVersion(text)
}

func compareVersions(a, b string) (int, error) {
	return reqcheck.CompareVersions(a, b)
}

func formatRequirementLabel(req automation.Requirement) string {
	return reqcheck.FormatRequirementLabel(req)
}

func installHint(req automation.Requirement) string {
	return reqcheck.InstallHintFor(req)
}
