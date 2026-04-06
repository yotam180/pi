package automation

import "strings"

// IsFilePath returns true if the value looks like a file path rather than
// inline script. The ext parameter is the file extension for the step type
// (e.g. ".sh", ".py", ".ts"). A value is a file path when it ends with ext,
// contains no newlines, and contains no spaces. Returns false when ext is empty.
func IsFilePath(value, ext string) bool {
	return ext != "" &&
		strings.HasSuffix(value, ext) &&
		!strings.Contains(value, "\n") &&
		!strings.Contains(value, " ")
}

// defaultFileExtensions maps each step type to the file extension its runner
// owns. Step types without file extensions (e.g. run:) are omitted.
// This is the static, compile-time source of truth for validation.
// The executor's SubprocessConfig.FileExt is the runtime equivalent.
var defaultFileExtensions = map[StepType]string{
	StepTypeBash:       ".sh",
	StepTypePython:     ".py",
	StepTypeTypeScript: ".ts",
}

// DefaultFileExtensions returns a copy of the default file extension map.
// Used by validation to detect file references without importing executor.
func DefaultFileExtensions() map[StepType]string {
	m := make(map[StepType]string, len(defaultFileExtensions))
	for k, v := range defaultFileExtensions {
		m[k] = v
	}
	return m
}

// parentShellStepTypes lists step types that support parent_shell: true.
var parentShellStepTypes = map[StepType]bool{
	StepTypeBash: true,
}

// StepTypeSupportsParentShell reports whether a step type supports
// parent_shell: true. This is the static, compile-time source of truth
// for validation. The executor's Registry.StepTypeSupportsParentShell()
// is the runtime equivalent that queries the actual runner.
func StepTypeSupportsParentShell(stepType StepType) bool {
	return parentShellStepTypes[stepType]
}
