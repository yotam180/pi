package automation

import (
	"strings"
	"testing"
)

func TestLoad_RequiresRuntimeBare(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "req.yaml", `
name: test
description: Test
requires:
  - python
  - node
steps:
  - bash: echo hello
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(a.Requires) != 2 {
		t.Fatalf("expected 2 requirements, got %d", len(a.Requires))
	}
	if a.Requires[0].Name != "python" || a.Requires[0].Kind != RequirementRuntime || a.Requires[0].MinVersion != "" {
		t.Errorf("req[0] = %+v, want python runtime with no version", a.Requires[0])
	}
	if a.Requires[1].Name != "node" || a.Requires[1].Kind != RequirementRuntime || a.Requires[1].MinVersion != "" {
		t.Errorf("req[1] = %+v, want node runtime with no version", a.Requires[1])
	}
}

func TestLoad_RequiresRuntimeWithVersion(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "req.yaml", `
name: test
description: Test
requires:
  - python >= 3.11
  - node >= 18
steps:
  - bash: echo hello
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(a.Requires) != 2 {
		t.Fatalf("expected 2 requirements, got %d", len(a.Requires))
	}
	if a.Requires[0].Name != "python" || a.Requires[0].Kind != RequirementRuntime || a.Requires[0].MinVersion != "3.11" {
		t.Errorf("req[0] = %+v, want python >= 3.11", a.Requires[0])
	}
	if a.Requires[1].Name != "node" || a.Requires[1].Kind != RequirementRuntime || a.Requires[1].MinVersion != "18" {
		t.Errorf("req[1] = %+v, want node >= 18", a.Requires[1])
	}
}

func TestLoad_RequiresCommandBare(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "req.yaml", `
name: test
description: Test
requires:
  - command: docker
  - command: jq
steps:
  - bash: echo hello
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(a.Requires) != 2 {
		t.Fatalf("expected 2 requirements, got %d", len(a.Requires))
	}
	if a.Requires[0].Name != "docker" || a.Requires[0].Kind != RequirementCommand || a.Requires[0].MinVersion != "" {
		t.Errorf("req[0] = %+v, want command:docker no version", a.Requires[0])
	}
	if a.Requires[1].Name != "jq" || a.Requires[1].Kind != RequirementCommand || a.Requires[1].MinVersion != "" {
		t.Errorf("req[1] = %+v, want command:jq no version", a.Requires[1])
	}
}

func TestLoad_RequiresCommandWithVersion(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "req.yaml", `
name: test
description: Test
requires:
  - command: kubectl >= 1.28
steps:
  - bash: echo hello
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(a.Requires) != 1 {
		t.Fatalf("expected 1 requirement, got %d", len(a.Requires))
	}
	if a.Requires[0].Name != "kubectl" || a.Requires[0].Kind != RequirementCommand || a.Requires[0].MinVersion != "1.28" {
		t.Errorf("req[0] = %+v, want command:kubectl >= 1.28", a.Requires[0])
	}
}

func TestLoad_RequiresMixed(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "req.yaml", `
name: test
description: Test
requires:
  - python >= 3.11
  - command: docker
  - command: jq
steps:
  - bash: echo hello
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(a.Requires) != 3 {
		t.Fatalf("expected 3 requirements, got %d", len(a.Requires))
	}
	if a.Requires[0].Kind != RequirementRuntime {
		t.Errorf("req[0].Kind = %q, want runtime", a.Requires[0].Kind)
	}
	if a.Requires[1].Kind != RequirementCommand {
		t.Errorf("req[1].Kind = %q, want command", a.Requires[1].Kind)
	}
	if a.Requires[2].Kind != RequirementCommand {
		t.Errorf("req[2].Kind = %q, want command", a.Requires[2].Kind)
	}
}

func TestLoad_RequiresThreePartVersion(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "req.yaml", `
name: test
description: Test
requires:
  - python >= 3.11.2
  - command: kubectl >= 1.28.0
steps:
  - bash: echo hello
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Requires[0].MinVersion != "3.11.2" {
		t.Errorf("req[0].MinVersion = %q, want 3.11.2", a.Requires[0].MinVersion)
	}
	if a.Requires[1].MinVersion != "1.28.0" {
		t.Errorf("req[1].MinVersion = %q, want 1.28.0", a.Requires[1].MinVersion)
	}
}

func TestLoad_RequiresNoBlock(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "req.yaml", `
name: test
steps:
  - bash: echo hello
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(a.Requires) != 0 {
		t.Errorf("expected 0 requirements, got %d", len(a.Requires))
	}
}

func TestLoad_RequiresOnInstaller(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "req.yaml", `
name: install-tool
description: Install tool
requires:
  - command: curl
install:
  test: command -v tool
  run: curl install.sh | sh
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !a.IsInstaller() {
		t.Fatal("expected installer automation")
	}
	if len(a.Requires) != 1 {
		t.Fatalf("expected 1 requirement, got %d", len(a.Requires))
	}
	if a.Requires[0].Name != "curl" || a.Requires[0].Kind != RequirementCommand {
		t.Errorf("req[0] = %+v, want command:curl", a.Requires[0])
	}
}

func TestLoad_RequiresUnknownRuntime(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "req.yaml", `
name: test
requires:
  - ruby >= 3.0
steps:
  - bash: echo hello
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for unknown runtime")
	}
	if !strings.Contains(err.Error(), "unknown runtime") {
		t.Errorf("expected 'unknown runtime' error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "command:") {
		t.Errorf("expected hint about command:, got: %v", err)
	}
}

func TestLoad_RequiresEmptyEntry(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "req.yaml", `
name: test
requires:
  - ""
steps:
  - bash: echo hello
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for empty requires entry")
	}
	if !strings.Contains(err.Error(), "cannot be empty") {
		t.Errorf("expected 'cannot be empty' error, got: %v", err)
	}
}

func TestLoad_RequiresEmptyCommand(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "req.yaml", `
name: test
requires:
  - command: ""
steps:
  - bash: echo hello
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for empty command value")
	}
	if !strings.Contains(err.Error(), "cannot be empty") {
		t.Errorf("expected 'cannot be empty' error, got: %v", err)
	}
}

func TestLoad_RequiresBadVersionSyntax(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "req.yaml", `
name: test
requires:
  - python >= abc
steps:
  - bash: echo hello
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for bad version syntax")
	}
	if !strings.Contains(err.Error(), "non-numeric") {
		t.Errorf("expected non-numeric error, got: %v", err)
	}
}

func TestLoad_RequiresMissingVersionAfterGte(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "req.yaml", `
name: test
requires:
  - python >=
steps:
  - bash: echo hello
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing version after >=")
	}
	if !strings.Contains(err.Error(), "missing version") {
		t.Errorf("expected 'missing version' error, got: %v", err)
	}
}

func TestLoad_RequiresInvalidMappingKey(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "req.yaml", `
name: test
requires:
  - runtime: python
steps:
  - bash: echo hello
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for unknown mapping key")
	}
	if !strings.Contains(err.Error(), "unknown key") {
		t.Errorf("expected 'unknown key' error, got: %v", err)
	}
}

func TestLoad_RequiresVersionWithEmptyComponent(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "req.yaml", `
name: test
requires:
  - python >= 3..11
steps:
  - bash: echo hello
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for version with empty component")
	}
	if !strings.Contains(err.Error(), "empty component") {
		t.Errorf("expected 'empty component' error, got: %v", err)
	}
}

func TestLoad_RequiresScalarWithSpaces(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "req.yaml", `
name: test
requires:
  - python something
steps:
  - bash: echo hello
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
	if !strings.Contains(err.Error(), "invalid format") {
		t.Errorf("expected 'invalid format' error, got: %v", err)
	}
}

func TestLoad_RequiresCommandVersionBadSyntax(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "req.yaml", `
name: test
requires:
  - command: kubectl >= x.y
steps:
  - bash: echo hello
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for bad command version")
	}
	if !strings.Contains(err.Error(), "non-numeric") {
		t.Errorf("expected non-numeric error, got: %v", err)
	}
}

func TestLoadFromBytes_RequiresBlock(t *testing.T) {
	data := []byte(`
name: test
description: Test
requires:
  - node >= 18
  - command: docker
steps:
  - bash: echo hello
`)

	a, err := LoadFromBytes(data, "builtin://test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(a.Requires) != 2 {
		t.Fatalf("expected 2 requirements, got %d", len(a.Requires))
	}
	if a.Requires[0].Name != "node" || a.Requires[0].MinVersion != "18" {
		t.Errorf("req[0] = %+v, want node >= 18", a.Requires[0])
	}
	if a.Requires[1].Name != "docker" || a.Requires[1].Kind != RequirementCommand {
		t.Errorf("req[1] = %+v, want command:docker", a.Requires[1])
	}
}

func TestParseNameVersion(t *testing.T) {
	tests := []struct {
		input       string
		wantName    string
		wantVersion string
		wantErr     bool
	}{
		{"docker", "docker", "", false},
		{"kubectl >= 1.28", "kubectl", "1.28", false},
		{"python >= 3.11.2", "python", "3.11.2", false},
		{"node >= 18", "node", "18", false},
		{">= 1.0", "", "", true},
		{"python >=", "", "", true},
		{"python >= abc", "", "", true},
		{"python >= 3..1", "", "", true},
		{"python something", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			name, version, err := parseNameVersion(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseNameVersion(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if name != tt.wantName {
					t.Errorf("name = %q, want %q", name, tt.wantName)
				}
				if version != tt.wantVersion {
					t.Errorf("version = %q, want %q", version, tt.wantVersion)
				}
			}
		})
	}
}

func TestValidateVersionString(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"3", false},
		{"3.11", false},
		{"3.11.2", false},
		{"18", false},
		{"1.28.0", false},
		{"abc", true},
		{"3..11", true},
		{"3.11.", true},
		{".3.11", true},
		{"3.11a", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			err := validateVersionString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateVersionString(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

