package executor

import (
	"github.com/vyper-tooling/pi/internal/automation"
	"github.com/vyper-tooling/pi/internal/conditions"
	"github.com/vyper-tooling/pi/internal/reqcheck"
	"github.com/vyper-tooling/pi/internal/runtimes"
)

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
		env = conditions.DefaultRuntimeEnv()
	}

	var failed []reqcheck.CheckResult
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
		return &reqcheck.ValidationError{
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
