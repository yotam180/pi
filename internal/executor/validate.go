package executor

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/vyper-tooling/pi/internal/automation"
)

// CheckResult holds the result of checking a single requirement.
type CheckResult struct {
	Requirement    automation.Requirement
	Satisfied      bool
	DetectedVersion string
	Error          string // human-readable reason when not satisfied
}

// ValidationError is returned when one or more requirements are not met.
type ValidationError struct {
	AutomationName string
	Results        []CheckResult
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("automation %q has unsatisfied requirements", e.AutomationName)
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
		label := formatRequirementLabel(r.Requirement)
		fmt.Fprintf(&b, "    %-25s %s\n", label, r.Error)
		if hint := installHint(r.Requirement); hint != "" {
			fmt.Fprintf(&b, "    %-25s → install: %s\n", "", hint)
		}
	}

	return b.String()
}

func formatRequirementLabel(req automation.Requirement) string {
	if req.MinVersion != "" {
		return fmt.Sprintf("%s >= %s", req.Name, req.MinVersion)
	}
	if req.Kind == automation.RequirementCommand {
		return fmt.Sprintf("command: %s", req.Name)
	}
	return req.Name
}

// ValidateRequirements checks all requirements on an automation.
// Returns nil if all are satisfied.
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
		result := checkRequirement(req, env)
		if !result.Satisfied {
			failed = append(failed, result)
		}
	}

	if len(failed) > 0 {
		return &ValidationError{
			AutomationName: a.Name,
			Results:        failed,
		}
	}

	return nil
}

// checkRequirement checks a single requirement against the runtime environment.
func checkRequirement(req automation.Requirement, env *RuntimeEnv) CheckResult {
	cmdName := req.Name
	if req.Kind == automation.RequirementRuntime {
		cmdName = runtimeCommand(req.Name)
	}

	_, err := env.LookPath(cmdName)
	if err != nil {
		return CheckResult{
			Requirement: req,
			Satisfied:   false,
			Error:       "not found",
		}
	}

	if req.MinVersion == "" {
		return CheckResult{
			Requirement: req,
			Satisfied:   true,
		}
	}

	detected := detectVersion(cmdName, env)
	if detected == "" {
		return CheckResult{
			Requirement: req,
			Satisfied:   false,
			Error:       fmt.Sprintf("installed but unable to detect version (tried %s --version)", cmdName),
		}
	}

	cmp, err := compareVersions(detected, req.MinVersion)
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

// runtimeCommand maps runtime names to their CLI command.
func runtimeCommand(name string) string {
	switch name {
	case "python":
		return "python3"
	default:
		return name
	}
}

// detectVersion runs `<cmd> --version` and extracts a semver-like version string.
func detectVersion(cmd string, env *RuntimeEnv) string {
	if env.ExecOutput == nil {
		return detectVersionExec(cmd)
	}
	raw := env.ExecOutput(cmd, "--version")
	return extractVersion(raw)
}

// detectVersionExec is the real implementation that runs the command.
func detectVersionExec(cmdName string) string {
	cmd := exec.Command(cmdName, "--version")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return ""
	}
	return extractVersion(out.String())
}

// versionRegex matches dot-separated numeric version strings (e.g. 3.13.0, 20.11.0, 1.7.1).
var versionRegex = regexp.MustCompile(`(\d+(?:\.\d+)+)`)

// extractVersion pulls the first semver-like version from text.
// Handles formats like: "Python 3.13.0", "v20.11.0", "jq-1.7.1", "node v22.0.0".
func extractVersion(text string) string {
	match := versionRegex.FindString(text)
	return match
}

// compareVersions compares two dot-separated numeric version strings.
// Returns -1, 0, or 1 (like strings.Compare semantics).
// "3.9.7" vs "3.11" → -1
// "3.13.0" vs "3.11" → 1
func compareVersions(a, b string) (int, error) {
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

// installHints maps tool names to their install commands.
var installHints = map[string]string{
	"python":     "brew install python3  or  https://www.python.org/downloads/",
	"node":       "brew install node  or  https://nodejs.org/",
	"docker":     "brew install --cask docker  or  https://docs.docker.com/get-docker/",
	"jq":         "brew install jq",
	"kubectl":    "brew install kubectl",
	"helm":       "brew install helm",
	"tsx":        "npm install -g tsx",
	"git":        "brew install git",
	"curl":       "brew install curl",
	"wget":       "brew install wget",
	"make":       "xcode-select --install  (macOS)  or  apt install build-essential",
	"mise":       "curl https://mise.run | sh",
	"uv":         "curl -LsSf https://astral.sh/uv/install.sh | sh",
}

// installHint returns a human-readable install hint for a requirement.
func installHint(req automation.Requirement) string {
	if hint, ok := installHints[req.Name]; ok {
		return hint
	}
	return ""
}
