// Package reqcheck provides runtime requirement checking for automations.
// It validates that required tools and runtimes are installed and satisfy
// version constraints, formats user-facing error output, and provides
// install hints for common tools.
package reqcheck

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/vyper-tooling/pi/internal/automation"
	"github.com/vyper-tooling/pi/internal/conditions"
	"github.com/vyper-tooling/pi/internal/runtimeinfo"
	"github.com/vyper-tooling/pi/internal/tools"
)

// CheckResult holds the result of checking a single requirement.
type CheckResult struct {
	Requirement     automation.Requirement
	Satisfied       bool
	DetectedVersion string
	Error           string
}

// ValidationError is returned when one or more requirements are not met.
type ValidationError struct {
	AutomationName string
	Results        []CheckResult
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("automation %q has unsatisfied requirements", e.AutomationName)
}

// CheckRequirement checks a single requirement against the runtime environment.
func CheckRequirement(req automation.Requirement, env *conditions.RuntimeEnv) CheckResult {
	return checkImpl(req, env, false)
}

// CheckRequirementForDoctor checks a single requirement, always detecting the
// version even when no minimum version constraint is set. This is used by
// `pi doctor` to show the detected version for satisfied requirements.
func CheckRequirementForDoctor(req automation.Requirement, env *conditions.RuntimeEnv) CheckResult {
	return checkImpl(req, env, true)
}

// checkImpl is the shared implementation for requirement checking.
// When alwaysDetectVersion is true, version detection runs even without a
// min-version constraint (used by pi doctor to display detected versions).
func checkImpl(req automation.Requirement, env *conditions.RuntimeEnv, alwaysDetectVersion bool) CheckResult {
	cmdName := req.Name
	if req.Kind == automation.RequirementRuntime {
		cmdName = RuntimeCommand(req.Name)
	}

	_, err := env.LookPath(cmdName)
	if err != nil {
		return CheckResult{
			Requirement: req,
			Satisfied:   false,
			Error:       "not found",
		}
	}

	needVersion := req.MinVersion != "" || alwaysDetectVersion
	if !needVersion {
		return CheckResult{
			Requirement: req,
			Satisfied:   true,
		}
	}

	detected := DetectVersion(cmdName, env)

	if req.MinVersion == "" {
		return CheckResult{
			Requirement:     req,
			Satisfied:       true,
			DetectedVersion: detected,
		}
	}

	if detected == "" {
		return CheckResult{
			Requirement: req,
			Satisfied:   false,
			Error:       fmt.Sprintf("installed but unable to detect version (tried %s --version)", cmdName),
		}
	}

	cmp, err := CompareVersions(detected, req.MinVersion)
	if err != nil {
		return CheckResult{
			Requirement:     req,
			Satisfied:       false,
			DetectedVersion: detected,
			Error:           fmt.Sprintf("version parse error: %v", err),
		}
	}

	if cmp < 0 {
		return CheckResult{
			Requirement:     req,
			Satisfied:       false,
			DetectedVersion: detected,
			Error:           fmt.Sprintf("found %s but need >= %s", detected, req.MinVersion),
		}
	}

	return CheckResult{
		Requirement:     req,
		Satisfied:       true,
		DetectedVersion: detected,
	}
}

// RuntimeCommand maps runtime names to their CLI binary name.
func RuntimeCommand(name string) string {
	return runtimeinfo.Binary(name)
}

// DetectVersion runs `<cmd> --version` and extracts a semver-like version string.
// Falls back to `<cmd> version` if `--version` produces no version.
func DetectVersion(cmd string, env *conditions.RuntimeEnv) string {
	if env.ExecOutput == nil {
		return detectVersionExec(cmd)
	}
	raw := env.ExecOutput(cmd, "--version")
	if v := ExtractVersion(raw); v != "" {
		return v
	}
	raw = env.ExecOutput(cmd, "version")
	return ExtractVersion(raw)
}

// detectVersionExec is the real implementation that runs the command.
func detectVersionExec(cmdName string) string {
	cmd := exec.Command(cmdName, "--version")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err == nil {
		if v := ExtractVersion(out.String()); v != "" {
			return v
		}
	}

	out.Reset()
	cmd = exec.Command(cmdName, "version")
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err == nil {
		return ExtractVersion(out.String())
	}
	return ""
}

// versionRegex matches dot-separated numeric version strings (e.g. 3.13.0, 20.11.0, 1.7.1).
var versionRegex = regexp.MustCompile(`(\d+(?:\.\d+)+)`)

// ExtractVersion pulls the first semver-like version from text.
// Handles formats like: "Python 3.13.0", "v20.11.0", "jq-1.7.1", "node v22.0.0".
func ExtractVersion(text string) string {
	return versionRegex.FindString(text)
}

// CompareVersions compares two dot-separated numeric version strings.
// Returns -1, 0, or 1 (like strings.Compare semantics).
func CompareVersions(a, b string) (int, error) {
	aParts, err := parseVersionParts(a)
	if err != nil {
		return 0, fmt.Errorf("version %q: %w", a, err)
	}
	bParts, err := parseVersionParts(b)
	if err != nil {
		return 0, fmt.Errorf("version %q: %w", b, err)
	}

	maxLen := len(aParts)
	if len(bParts) > maxLen {
		maxLen = len(bParts)
	}

	for i := 0; i < maxLen; i++ {
		av := 0
		if i < len(aParts) {
			av = aParts[i]
		}
		bv := 0
		if i < len(bParts) {
			bv = bParts[i]
		}

		if av < bv {
			return -1, nil
		}
		if av > bv {
			return 1, nil
		}
	}

	return 0, nil
}

func parseVersionParts(v string) ([]int, error) {
	parts := strings.Split(v, ".")
	result := make([]int, len(parts))
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("non-numeric component %q", p)
		}
		result[i] = n
	}
	return result, nil
}

// FormatValidationError produces the user-facing error table for missing requirements.
func FormatValidationError(ve *ValidationError) string {
	var b strings.Builder

	fmt.Fprintf(&b, "✗ pi run %s\n\n", ve.AutomationName)
	fmt.Fprintf(&b, "  Missing requirements:\n")

	for _, r := range ve.Results {
		if r.Satisfied {
			continue
		}
		label := FormatRequirementLabel(r.Requirement)
		fmt.Fprintf(&b, "    %-25s %s\n", label, r.Error)
		if hint := InstallHintFor(r.Requirement); hint != "" {
			fmt.Fprintf(&b, "    %-25s → install: %s\n", "", hint)
		}
	}

	return b.String()
}

// FormatRequirementLabel formats a requirement for display.
func FormatRequirementLabel(req automation.Requirement) string {
	if req.MinVersion != "" {
		return fmt.Sprintf("%s >= %s", req.Name, req.MinVersion)
	}
	if req.Kind == automation.RequirementCommand {
		return fmt.Sprintf("command: %s", req.Name)
	}
	return req.Name
}

// InstallHintFor returns a human-readable install hint for a requirement.
func InstallHintFor(req automation.Requirement) string {
	return tools.InstallHintFor(req.Name)
}
