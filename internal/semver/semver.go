package semver

import (
	"fmt"
	"strings"

	sv "github.com/Masterminds/semver/v3"
)

// Satisfies checks whether a version string satisfies a constraint expression.
// It normalises incomplete versions (e.g. "22" → "22.0.0") and supports the
// constraint syntax from Masterminds/semver: exact ("22"), >=, ^, ~, ranges.
//
// When the constraint is a non-semver channel name (e.g. "stable", "nightly",
// "beta"), any valid version is accepted. This supports tools like Rust's
// rustup that use channel names alongside version numbers.
//
// Returns nil if satisfied, or an error describing the mismatch.
func Satisfies(version, constraint string) error {
	version = strings.TrimPrefix(version, "v")
	constraint = strings.TrimPrefix(constraint, "v")

	normVersion, err := normalise(version)
	if err != nil {
		return fmt.Errorf("invalid version %q: %w", version, err)
	}

	v, err := sv.NewVersion(normVersion)
	if err != nil {
		return fmt.Errorf("invalid version %q: %w", version, err)
	}

	if isChannelName(constraint) {
		return nil
	}

	c, err := parseConstraint(constraint)
	if err != nil {
		return fmt.Errorf("invalid constraint %q: %w", constraint, err)
	}

	if !c.Check(v) {
		return fmt.Errorf("%s does not satisfy %s", version, constraint)
	}
	return nil
}

// isChannelName returns true if the constraint is a non-semver channel name
// (e.g. "stable", "nightly", "beta"). These are purely alphabetic strings with
// no digits, dots, or semver operators. When a channel name is used as a
// constraint, any valid installed version satisfies it.
func isChannelName(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < 'a' || r > 'z' {
			return false
		}
	}
	return true
}

// parseConstraint handles the constraint string, normalising bare major/minor
// numbers into caret constraints (e.g. "22" → "^22.0.0", "22.3" → "^22.3.0").
func parseConstraint(s string) (*sv.Constraints, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("empty constraint")
	}

	// If the string starts with an operator or contains spaces (range), pass
	// it through to Masterminds directly after normalising version parts.
	if hasOperator(s) {
		normalised := normaliseConstraintParts(s)
		return sv.NewConstraint(normalised)
	}

	// Bare version: "22" or "22.3" or "22.3.1" — treat as caret (compatible-with).
	norm, err := normalise(s)
	if err != nil {
		return nil, err
	}
	return sv.NewConstraint("^" + norm)
}

// hasOperator returns true if the constraint string starts with a comparison
// operator or contains one after spaces (for ranges like ">= 1.0, < 2.0").
func hasOperator(s string) bool {
	s = strings.TrimSpace(s)
	for _, op := range []string{">=", "<=", "!=", ">", "<", "~", "^", "="} {
		if strings.HasPrefix(s, op) {
			return true
		}
	}
	return false
}

// normaliseConstraintParts normalises version numbers inside constraint expressions.
// For example, ">= 20" becomes ">= 20.0.0".
func normaliseConstraintParts(s string) string {
	parts := strings.Split(s, ",")
	for i, part := range parts {
		part = strings.TrimSpace(part)
		op, ver := splitOperator(part)
		if norm, err := normalise(ver); err == nil {
			parts[i] = op + norm
		}
	}
	return strings.Join(parts, ", ")
}

// splitOperator separates the operator prefix from the version string.
func splitOperator(s string) (string, string) {
	s = strings.TrimSpace(s)
	for _, op := range []string{">=", "<=", "!=", ">", "<", "~", "^", "="} {
		if strings.HasPrefix(s, op) {
			return op + " ", strings.TrimSpace(strings.TrimPrefix(s, op))
		}
	}
	return "", s
}

// normalise pads a version string to three components: "22" → "22.0.0", "22.3" → "22.3.0".
func normalise(s string) (string, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "v")
	if s == "" {
		return "", fmt.Errorf("empty version")
	}
	parts := strings.Split(s, ".")
	if len(parts) > 3 {
		return "", fmt.Errorf("too many version components in %q", s)
	}
	for len(parts) < 3 {
		parts = append(parts, "0")
	}
	return strings.Join(parts, "."), nil
}
