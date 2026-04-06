package runtimeinfo

import (
	"strings"
	"testing"
)

func TestFind_KnownRuntime(t *testing.T) {
	for _, rt := range Runtimes {
		d := Find(rt.Name)
		if d == nil {
			t.Errorf("Find(%q) returned nil", rt.Name)
			continue
		}
		if d.Name != rt.Name {
			t.Errorf("Find(%q).Name = %q, want %q", rt.Name, d.Name, rt.Name)
		}
	}
}

func TestFind_UnknownRuntime(t *testing.T) {
	d := Find("ruby")
	if d != nil {
		t.Errorf("Find(\"ruby\") = %+v, want nil", d)
	}
}

func TestKnownNames(t *testing.T) {
	names := KnownNames()
	for _, rt := range Runtimes {
		if !names[rt.Name] {
			t.Errorf("KnownNames() missing %q", rt.Name)
		}
	}
	if names["ruby"] {
		t.Error("KnownNames() should not contain \"ruby\"")
	}
	if len(names) != len(Runtimes) {
		t.Errorf("KnownNames() has %d entries, want %d", len(names), len(Runtimes))
	}
}

func TestSortedNames(t *testing.T) {
	s := SortedNames()
	for _, name := range []string{"go", "node", "python", "rust"} {
		if !strings.Contains(s, name) {
			t.Errorf("SortedNames() = %q, missing %q", s, name)
		}
	}
	if !strings.Contains(s, ", ") {
		t.Errorf("SortedNames() should be comma-separated, got %q", s)
	}
	// Verify sorted order
	if s != "go, node, python, rust" {
		t.Errorf("SortedNames() = %q, want \"go, node, python, rust\"", s)
	}
}

func TestBinary_KnownRuntimes(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"python", "python3"},
		{"node", "node"},
		{"go", "go"},
		{"rust", "rustc"},
	}
	for _, tt := range tests {
		if got := Binary(tt.name); got != tt.want {
			t.Errorf("Binary(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestBinary_UnknownRuntime(t *testing.T) {
	if got := Binary("ruby"); got != "ruby" {
		t.Errorf("Binary(\"ruby\") = %q, want \"ruby\"", got)
	}
}

func TestDefaultVersion_KnownRuntimes(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"python", "3.13"},
		{"node", "20"},
		{"go", "1.23"},
		{"rust", "stable"},
	}
	for _, tt := range tests {
		if got := DefaultVersion(tt.name); got != tt.want {
			t.Errorf("DefaultVersion(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestDefaultVersion_UnknownRuntime(t *testing.T) {
	if got := DefaultVersion("ruby"); got != "latest" {
		t.Errorf("DefaultVersion(\"ruby\") = %q, want \"latest\"", got)
	}
}

func TestRuntimes_AllHaveRequiredFields(t *testing.T) {
	for _, rt := range Runtimes {
		if rt.Name == "" {
			t.Error("runtime with empty Name")
		}
		if rt.Binary == "" {
			t.Errorf("runtime %q has empty Binary", rt.Name)
		}
		if rt.DefaultVersion == "" {
			t.Errorf("runtime %q has empty DefaultVersion", rt.Name)
		}
		if rt.InstallHint == "" {
			t.Errorf("runtime %q has empty InstallHint", rt.Name)
		}
	}
}

func TestRuntimes_NoDuplicateNames(t *testing.T) {
	seen := make(map[string]bool)
	for _, rt := range Runtimes {
		if seen[rt.Name] {
			t.Errorf("duplicate runtime name %q", rt.Name)
		}
		seen[rt.Name] = true
	}
}
