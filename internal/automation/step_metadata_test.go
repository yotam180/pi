package automation

import "testing"

func TestIsFilePath(t *testing.T) {
	tests := []struct {
		value string
		ext   string
		want  bool
	}{
		{"script.sh", ".sh", true},
		{"path/to/script.sh", ".sh", true},
		{"transform.py", ".py", true},
		{"app.ts", ".ts", true},
		{"echo hello", ".sh", false},
		{"echo hello.sh world", ".sh", false},
		{"line1\nline2.sh", ".sh", false},
		{"script.py", ".sh", false},
		{"script.sh", "", false},
		{"", ".sh", false},
		{"", "", false},
	}
	for _, tt := range tests {
		got := IsFilePath(tt.value, tt.ext)
		if got != tt.want {
			t.Errorf("IsFilePath(%q, %q) = %v, want %v", tt.value, tt.ext, got, tt.want)
		}
	}
}

func TestDefaultFileExtensions(t *testing.T) {
	exts := DefaultFileExtensions()

	if exts[StepTypeBash] != ".sh" {
		t.Errorf("bash ext = %q, want .sh", exts[StepTypeBash])
	}
	if exts[StepTypePython] != ".py" {
		t.Errorf("python ext = %q, want .py", exts[StepTypePython])
	}
	if exts[StepTypeTypeScript] != ".ts" {
		t.Errorf("typescript ext = %q, want .ts", exts[StepTypeTypeScript])
	}
	if _, ok := exts[StepTypeRun]; ok {
		t.Error("run step type should not have a file extension")
	}

	exts[StepTypeBash] = ".modified"
	fresh := DefaultFileExtensions()
	if fresh[StepTypeBash] != ".sh" {
		t.Error("DefaultFileExtensions should return a copy, not the internal map")
	}
}

func TestStepTypeSupportsParentShell(t *testing.T) {
	tests := []struct {
		stepType StepType
		want     bool
	}{
		{StepTypeBash, true},
		{StepTypePython, false},
		{StepTypeTypeScript, false},
		{StepTypeRun, false},
		{StepType("unknown"), false},
	}
	for _, tt := range tests {
		got := StepTypeSupportsParentShell(tt.stepType)
		if got != tt.want {
			t.Errorf("StepTypeSupportsParentShell(%q) = %v, want %v", tt.stepType, got, tt.want)
		}
	}
}

func TestDefaultFileExtensionsConsistency(t *testing.T) {
	exts := DefaultFileExtensions()
	for stepType, ext := range exts {
		if ext == "" {
			t.Errorf("step type %q has empty extension in DefaultFileExtensions", stepType)
		}
		if !stepType.IsValid() {
			t.Errorf("step type %q in DefaultFileExtensions is not a valid step type", stepType)
		}
	}
}
