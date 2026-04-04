package builtins

import (
	"strings"
	"testing"

	"github.com/vyper-tooling/pi/internal/automation"
)

func TestDiscover_FindsEmbeddedAutomations(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	if len(result.Automations) == 0 {
		t.Fatal("expected at least one built-in automation")
	}

	a, ok := result.Automations["hello"]
	if !ok {
		t.Fatal("expected to find built-in 'hello' automation")
	}

	if a.Name != "hello" {
		t.Errorf("expected name 'hello', got %q", a.Name)
	}

	if a.Description == "" {
		t.Error("expected non-empty description for hello automation")
	}

	if len(a.Steps) == 0 {
		t.Error("expected at least one step in hello automation")
	}
}

func TestDiscover_NamesAreSorted(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	names := result.Names()
	for i := 1; i < len(names); i++ {
		if names[i] < names[i-1] {
			t.Errorf("names not sorted: %q came before %q", names[i-1], names[i])
		}
	}
}

func TestDiscover_ResultIsUsableWithFind(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	a, err := result.Find("hello")
	if err != nil {
		t.Fatalf("Find('hello') returned error: %v", err)
	}

	if a.Name != "hello" {
		t.Errorf("expected name 'hello', got %q", a.Name)
	}
}

func TestDiscover_DockerAutomationsExist(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	dockerAutomations := []struct {
		name        string
		description string
	}{
		{"docker/up", "Start Docker Compose services in detached mode"},
		{"docker/down", "Stop and remove Docker Compose services"},
		{"docker/logs", "Stream Docker Compose service logs"},
	}

	for _, tc := range dockerAutomations {
		t.Run(tc.name, func(t *testing.T) {
			a, ok := result.Automations[tc.name]
			if !ok {
				t.Fatalf("expected to find built-in %q automation", tc.name)
			}

			if a.Name != tc.name {
				t.Errorf("expected name %q, got %q", tc.name, a.Name)
			}

			if a.Description != tc.description {
				t.Errorf("expected description %q, got %q", tc.description, a.Description)
			}

			if len(a.Steps) == 0 {
				t.Error("expected at least one step")
			}
		})
	}
}

func TestDiscover_DockerAutomationsAreResolvable(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	for _, name := range []string{"docker/up", "docker/down", "docker/logs"} {
		t.Run(name, func(t *testing.T) {
			a, err := result.Find(name)
			if err != nil {
				t.Fatalf("Find(%q) returned error: %v", name, err)
			}
			if a.Name != name {
				t.Errorf("expected name %q, got %q", name, a.Name)
			}
		})
	}
}

func TestDiscover_DockerAutomationsUseBashSteps(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	for _, name := range []string{"docker/up", "docker/down", "docker/logs"} {
		t.Run(name, func(t *testing.T) {
			a := result.Automations[name]
			if len(a.Steps) != 1 {
				t.Fatalf("expected 1 step, got %d", len(a.Steps))
			}
			step := a.Steps[0]
			if step.Type != automation.StepTypeBash {
				t.Errorf("expected bash step, got %q", step.Type)
			}
		})
	}
}

func TestDiscover_DockerAutomationsDetectComposeVersions(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	for _, name := range []string{"docker/up", "docker/down", "docker/logs"} {
		t.Run(name, func(t *testing.T) {
			a := result.Automations[name]
			script := a.Steps[0].Value
			if !strings.Contains(script, "docker compose") {
				t.Error("expected script to reference 'docker compose' (v2 plugin)")
			}
			if !strings.Contains(script, "docker-compose") {
				t.Error("expected script to reference 'docker-compose' (standalone fallback)")
			}
		})
	}
}

func TestDiscover_DockerUpForwardsArgs(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	a := result.Automations["docker/up"]
	script := a.Steps[0].Value
	if !strings.Contains(script, `"$@"`) {
		t.Error("expected docker/up script to forward args via \"$@\"")
	}
	if !strings.Contains(script, "up -d") {
		t.Error("expected docker/up script to include 'up -d'")
	}
}

func TestDiscover_DockerDownForwardsArgs(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	a := result.Automations["docker/down"]
	script := a.Steps[0].Value
	if !strings.Contains(script, `"$@"`) {
		t.Error("expected docker/down script to forward args via \"$@\"")
	}
	if !strings.Contains(script, "down") {
		t.Error("expected docker/down script to include 'down'")
	}
}

func TestDiscover_DockerLogsForwardsArgs(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	a := result.Automations["docker/logs"]
	script := a.Steps[0].Value
	if !strings.Contains(script, `"$@"`) {
		t.Error("expected docker/logs script to forward args via \"$@\"")
	}
	if !strings.Contains(script, "logs -f --tail 200") {
		t.Error("expected docker/logs script to include 'logs -f --tail 200'")
	}
}

func TestDiscover_InstallerAutomationsExist(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	installers := []struct {
		name        string
		description string
	}{
		{"install-homebrew", "Install Homebrew (macOS only)"},
		{"install-python", "Install Python at a specific version"},
		{"install-node", "Install Node.js at a specific version"},
		{"install-uv", "Install the uv Python package manager"},
		{"install-tsx", "Install tsx globally for TypeScript execution"},
	}

	for _, tc := range installers {
		t.Run(tc.name, func(t *testing.T) {
			a, ok := result.Automations[tc.name]
			if !ok {
				t.Fatalf("expected to find built-in %q automation", tc.name)
			}

			if a.Name != tc.name {
				t.Errorf("expected name %q, got %q", tc.name, a.Name)
			}

			if a.Description != tc.description {
				t.Errorf("expected description %q, got %q", tc.description, a.Description)
			}

			if len(a.Steps) == 0 {
				t.Error("expected at least one step")
			}
		})
	}
}

func TestDiscover_InstallerAutomationsAreResolvable(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	names := []string{"install-homebrew", "install-python", "install-node", "install-uv", "install-tsx"}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			a, err := result.Find(name)
			if err != nil {
				t.Fatalf("Find(%q) returned error: %v", name, err)
			}
			if a.Name != name {
				t.Errorf("expected name %q, got %q", name, a.Name)
			}
		})
	}
}

func TestDiscover_InstallerAutomationsUseBashSteps(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	names := []string{"install-homebrew", "install-python", "install-node", "install-uv", "install-tsx"}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			a := result.Automations[name]
			if len(a.Steps) != 1 {
				t.Fatalf("expected 1 step, got %d", len(a.Steps))
			}
			if a.Steps[0].Type != automation.StepTypeBash {
				t.Errorf("expected bash step, got %q", a.Steps[0].Type)
			}
		})
	}
}

func TestDiscover_InstallerAutomationsAreIdempotent(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	names := []string{"install-homebrew", "install-python", "install-node", "install-uv", "install-tsx"}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			a := result.Automations[name]
			script := a.Steps[0].Value
			if !strings.Contains(script, "command -v") && !strings.Contains(script, "which") {
				t.Error("expected script to check if tool is already installed (command -v or which)")
			}
			if !strings.Contains(script, "[already installed]") {
				t.Error("expected script to print '[already installed]' when tool exists")
			}
			if !strings.Contains(script, "[installed]") {
				t.Error("expected script to print '[installed]' after installing")
			}
		})
	}
}

func TestDiscover_InstallHomebrewHasCondition(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	a := result.Automations["install-homebrew"]
	if a.If != "os.macos" {
		t.Errorf("expected install-homebrew to have if: 'os.macos', got %q", a.If)
	}
}

func TestDiscover_InstallPythonAcceptsVersionInput(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	a := result.Automations["install-python"]
	if len(a.Inputs) == 0 {
		t.Fatal("expected install-python to have inputs")
	}
	spec, ok := a.Inputs["version"]
	if !ok {
		t.Fatal("expected install-python to have a 'version' input")
	}
	if !spec.IsRequired() {
		t.Error("expected 'version' input to be required")
	}
	if spec.Description == "" {
		t.Error("expected 'version' input to have a description")
	}

	script := a.Steps[0].Value
	if !strings.Contains(script, "PI_INPUT_VERSION") {
		t.Error("expected script to use PI_INPUT_VERSION env var")
	}
}

func TestDiscover_InstallNodeAcceptsVersionInput(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	a := result.Automations["install-node"]
	if len(a.Inputs) == 0 {
		t.Fatal("expected install-node to have inputs")
	}
	spec, ok := a.Inputs["version"]
	if !ok {
		t.Fatal("expected install-node to have a 'version' input")
	}
	if !spec.IsRequired() {
		t.Error("expected 'version' input to be required")
	}
	if spec.Description == "" {
		t.Error("expected 'version' input to have a description")
	}

	script := a.Steps[0].Value
	if !strings.Contains(script, "PI_INPUT_VERSION") {
		t.Error("expected script to use PI_INPUT_VERSION env var")
	}
}

func TestDiscover_InstallUvUsesOfficialInstaller(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	a := result.Automations["install-uv"]
	script := a.Steps[0].Value
	if !strings.Contains(script, "astral.sh/uv/install.sh") {
		t.Error("expected install-uv to use the official astral.sh installer")
	}
}

func TestDiscover_InstallTsxUsesNpm(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	a := result.Automations["install-tsx"]
	script := a.Steps[0].Value
	if !strings.Contains(script, "npm install -g tsx") {
		t.Error("expected install-tsx to use 'npm install -g tsx'")
	}
}

func TestDiscover_InstallPythonUsesMiseAndBrew(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	a := result.Automations["install-python"]
	script := a.Steps[0].Value
	if !strings.Contains(script, "mise") {
		t.Error("expected install-python to try mise first")
	}
	if !strings.Contains(script, "brew") {
		t.Error("expected install-python to fall back to brew")
	}
}

func TestDiscover_InstallNodeUsesMiseAndBrew(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	a := result.Automations["install-node"]
	script := a.Steps[0].Value
	if !strings.Contains(script, "mise") {
		t.Error("expected install-node to try mise first")
	}
	if !strings.Contains(script, "brew") {
		t.Error("expected install-node to fall back to brew")
	}
}

func TestDiscover_InstallerAutomationsNoInputs(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	noInputs := []string{"install-homebrew", "install-uv", "install-tsx"}
	for _, name := range noInputs {
		t.Run(name, func(t *testing.T) {
			a := result.Automations[name]
			if len(a.Inputs) != 0 {
				t.Errorf("expected %s to have no inputs, got %d", name, len(a.Inputs))
			}
		})
	}
}
