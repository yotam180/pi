package interpolation

import (
	"fmt"
	"strings"
)

// OutputTracker records step outputs and provides indexed/last-output lookups.
// It supports save/restore for scoped resets (e.g., installer phases that
// should not pollute the caller's output history).
type OutputTracker struct {
	outputs []string
}

// Record appends a step's captured stdout (trimmed) to the outputs list.
func (t *OutputTracker) Record(output string) {
	t.outputs = append(t.outputs, strings.TrimSpace(output))
}

// Last returns the trimmed stdout of the most recently recorded step,
// or "" if no step has produced output yet.
func (t *OutputTracker) Last() string {
	if len(t.outputs) == 0 {
		return ""
	}
	return t.outputs[len(t.outputs)-1]
}

// Get returns the output at the given 0-based index, or "" if out of range.
// The second return value reports whether the index was valid.
func (t *OutputTracker) Get(index int) (string, bool) {
	if index < 0 || index >= len(t.outputs) {
		return "", false
	}
	return t.outputs[index], true
}

// Len returns the number of recorded outputs.
func (t *OutputTracker) Len() int {
	return len(t.outputs)
}

// Snapshot captures the current outputs so they can be restored later via
// Restore. This is used to scope output tracking within sub-automations
// (e.g., install phases) without losing the caller's accumulated outputs.
func (t *OutputTracker) Snapshot() []string {
	if t.outputs == nil {
		return nil
	}
	cp := make([]string, len(t.outputs))
	copy(cp, t.outputs)
	return cp
}

// Restore replaces the current outputs with a previously captured snapshot.
func (t *OutputTracker) Restore(saved []string) {
	t.outputs = saved
}

// Reset clears all recorded outputs.
func (t *OutputTracker) Reset() {
	t.outputs = nil
}

// ResolveValue replaces "outputs.last", "outputs.<N>", and "inputs.<name>"
// references in a string value. Unrecognized values pass through unchanged.
func ResolveValue(v string, tracker *OutputTracker, inputEnv []string) string {
	if v == "outputs.last" {
		if tracker == nil {
			return ""
		}
		return tracker.Last()
	}
	if strings.HasPrefix(v, "outputs.") {
		suffix := strings.TrimPrefix(v, "outputs.")
		var n int
		if _, err := fmt.Sscanf(suffix, "%d", &n); err == nil {
			if tracker != nil {
				if val, ok := tracker.Get(n); ok {
					return val
				}
			}
		}
		return v
	}
	if strings.HasPrefix(v, "inputs.") {
		name := strings.TrimPrefix(v, "inputs.")
		envKey := "PI_IN_" + strings.ToUpper(strings.ReplaceAll(name, "-", "_"))
		for _, entry := range inputEnv {
			if strings.HasPrefix(entry, envKey+"=") {
				return entry[len(envKey)+1:]
			}
		}
		return v
	}
	return v
}

// ResolveEnv applies ResolveValue to each value in an env map. Returns the
// original map unchanged if no values contain interpolation references.
func ResolveEnv(env map[string]string, tracker *OutputTracker, inputEnv []string) map[string]string {
	if len(env) == 0 {
		return env
	}
	result := make(map[string]string, len(env))
	changed := false
	for k, v := range env {
		resolved := ResolveValue(v, tracker, inputEnv)
		result[k] = resolved
		if resolved != v {
			changed = true
		}
	}
	if !changed {
		return env
	}
	return result
}

// ResolveWith resolves output and input references in with: values.
// Returns the original map when empty/nil.
func ResolveWith(with map[string]string, tracker *OutputTracker, inputEnv []string) map[string]string {
	if len(with) == 0 {
		return with
	}
	result := make(map[string]string, len(with))
	for k, v := range with {
		result[k] = ResolveValue(v, tracker, inputEnv)
	}
	return result
}
