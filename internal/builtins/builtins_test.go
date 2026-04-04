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

			if !a.IsInstaller() {
				t.Error("expected automation to use install: block")
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

func TestDiscover_InstallerAutomationsHaveInstallBlock(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	names := []string{"install-homebrew", "install-python", "install-node", "install-uv", "install-tsx"}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			a := result.Automations[name]
			if !a.IsInstaller() {
				t.Fatal("expected automation to use install: block")
			}
			inst := a.Install
			if inst.Test.IsScalar && inst.Test.Scalar == "" {
				t.Error("expected non-empty test phase")
			}
			if !inst.Test.IsScalar && len(inst.Test.Steps) == 0 {
				t.Error("expected non-empty test phase steps")
			}
			if inst.Run.IsScalar && inst.Run.Scalar == "" {
				t.Error("expected non-empty run phase")
			}
			if !inst.Run.IsScalar && len(inst.Run.Steps) == 0 {
				t.Error("expected non-empty run phase steps")
			}
		})
	}
}

func TestDiscover_InstallerAutomationsHaveTestPhase(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	scalarInstallers := map[string]string{
		"install-homebrew": "command -v brew",
		"install-uv":       "command -v uv",
		"install-tsx":      "command -v tsx",
	}
	for name, expectedContent := range scalarInstallers {
		t.Run(name, func(t *testing.T) {
			a := result.Automations[name]
			if !a.Install.Test.IsScalar {
				t.Fatal("expected scalar test phase")
			}
			if !strings.Contains(a.Install.Test.Scalar, expectedContent) {
				t.Errorf("expected test to contain %q, got %q", expectedContent, a.Install.Test.Scalar)
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

	if !a.IsInstaller() {
		t.Fatal("expected install-python to be an installer")
	}
	testPhase := a.Install.Test
	if testPhase.IsScalar {
		if !strings.Contains(testPhase.Scalar, "PI_INPUT_VERSION") {
			t.Error("expected test phase to reference PI_INPUT_VERSION")
		}
	} else if len(testPhase.Steps) > 0 {
		found := false
		for _, s := range testPhase.Steps {
			if strings.Contains(s.Value, "PI_INPUT_VERSION") {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected test phase steps to reference PI_INPUT_VERSION")
		}
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

	if !a.IsInstaller() {
		t.Fatal("expected install-node to be an installer")
	}
	testPhase := a.Install.Test
	if testPhase.IsScalar {
		if !strings.Contains(testPhase.Scalar, "PI_INPUT_VERSION") {
			t.Error("expected test phase to reference PI_INPUT_VERSION")
		}
	} else if len(testPhase.Steps) > 0 {
		found := false
		for _, s := range testPhase.Steps {
			if strings.Contains(s.Value, "PI_INPUT_VERSION") {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected test phase steps to reference PI_INPUT_VERSION")
		}
	}
}

func TestDiscover_InstallUvUsesOfficialInstaller(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	a := result.Automations["install-uv"]
	if !a.IsInstaller() {
		t.Fatal("expected installer automation")
	}
	run := a.Install.Run
	if run.IsScalar {
		if !strings.Contains(run.Scalar, "astral.sh/uv/install.sh") {
			t.Error("expected install-uv run phase to use the official astral.sh installer")
		}
	}
}

func TestDiscover_InstallTsxUsesNpm(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	a := result.Automations["install-tsx"]
	if !a.IsInstaller() {
		t.Fatal("expected installer automation")
	}
	run := a.Install.Run
	if run.IsScalar {
		if !strings.Contains(run.Scalar, "npm install -g tsx") {
			t.Error("expected install-tsx run phase to use 'npm install -g tsx'")
		}
	}
}

func TestDiscover_InstallPythonUsesMiseAndBrew(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	a := result.Automations["install-python"]
	if !a.IsInstaller() {
		t.Fatal("expected installer automation")
	}
	run := a.Install.Run
	if run.IsScalar {
		t.Fatal("expected step list for install-python run phase")
	}
	foundMise, foundBrew := false, false
	for _, s := range run.Steps {
		if strings.Contains(s.Value, "mise") {
			foundMise = true
		}
		if strings.Contains(s.Value, "brew") {
			foundBrew = true
		}
	}
	if !foundMise {
		t.Error("expected install-python to try mise")
	}
	if !foundBrew {
		t.Error("expected install-python to fall back to brew")
	}
}

func TestDiscover_InstallNodeUsesMiseAndBrew(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	a := result.Automations["install-node"]
	if !a.IsInstaller() {
		t.Fatal("expected installer automation")
	}
	run := a.Install.Run
	if run.IsScalar {
		t.Fatal("expected step list for install-node run phase")
	}
	foundMise, foundBrew := false, false
	for _, s := range run.Steps {
		if strings.Contains(s.Value, "mise") {
			foundMise = true
		}
		if strings.Contains(s.Value, "brew") {
			foundBrew = true
		}
	}
	if !foundMise {
		t.Error("expected install-node to try mise")
	}
	if !foundBrew {
		t.Error("expected install-node to fall back to brew")
	}
}

func TestDiscover_CursorInstallExtensionsExists(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	a, ok := result.Automations["cursor/install-extensions"]
	if !ok {
		t.Fatal("expected to find built-in 'cursor/install-extensions' automation")
	}

	if a.Name != "cursor/install-extensions" {
		t.Errorf("expected name 'cursor/install-extensions', got %q", a.Name)
	}

	if a.Description != "Install missing Cursor extensions from a list" {
		t.Errorf("expected description 'Install missing Cursor extensions from a list', got %q", a.Description)
	}

	if len(a.Steps) == 0 {
		t.Error("expected at least one step")
	}
}

func TestDiscover_CursorInstallExtensionsIsResolvable(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	a, err := result.Find("cursor/install-extensions")
	if err != nil {
		t.Fatalf("Find('cursor/install-extensions') returned error: %v", err)
	}
	if a.Name != "cursor/install-extensions" {
		t.Errorf("expected name 'cursor/install-extensions', got %q", a.Name)
	}
}

func TestDiscover_CursorInstallExtensionsUsesBash(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	a := result.Automations["cursor/install-extensions"]
	if len(a.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(a.Steps))
	}
	if a.Steps[0].Type != automation.StepTypeBash {
		t.Errorf("expected bash step, got %q", a.Steps[0].Type)
	}
}

func TestDiscover_CursorInstallExtensionsAcceptsExtensionsInput(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	a := result.Automations["cursor/install-extensions"]
	if len(a.Inputs) == 0 {
		t.Fatal("expected cursor/install-extensions to have inputs")
	}
	spec, ok := a.Inputs["extensions"]
	if !ok {
		t.Fatal("expected cursor/install-extensions to have an 'extensions' input")
	}
	if !spec.IsRequired() {
		t.Error("expected 'extensions' input to be required")
	}
	if spec.Description == "" {
		t.Error("expected 'extensions' input to have a description")
	}
}

func TestDiscover_CursorInstallExtensionsUsesInput(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	a := result.Automations["cursor/install-extensions"]
	script := a.Steps[0].Value
	if !strings.Contains(script, "PI_INPUT_EXTENSIONS") {
		t.Error("expected script to use PI_INPUT_EXTENSIONS env var")
	}
}

func TestDiscover_CursorInstallExtensionsIsIdempotent(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	a := result.Automations["cursor/install-extensions"]
	script := a.Steps[0].Value
	if !strings.Contains(script, "cursor --list-extensions") {
		t.Error("expected script to check existing extensions via cursor --list-extensions")
	}
	if !strings.Contains(script, "[already installed]") {
		t.Error("expected script to print '[already installed]' when all extensions present")
	}
	if !strings.Contains(script, "[installed]") {
		t.Error("expected script to print '[installed]' after installing new extensions")
	}
}

func TestDiscover_CursorInstallExtensionsUsesCursorCLI(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	a := result.Automations["cursor/install-extensions"]
	script := a.Steps[0].Value
	if !strings.Contains(script, "cursor --install-extension") {
		t.Error("expected script to install extensions via 'cursor --install-extension'")
	}
}

func TestDiscover_GitInstallHooksExists(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	a, ok := result.Automations["git/install-hooks"]
	if !ok {
		t.Fatal("expected to find built-in 'git/install-hooks' automation")
	}

	if a.Name != "git/install-hooks" {
		t.Errorf("expected name 'git/install-hooks', got %q", a.Name)
	}

	if a.Description != "Install git hooks from a source directory" {
		t.Errorf("expected description 'Install git hooks from a source directory', got %q", a.Description)
	}

	if len(a.Steps) == 0 {
		t.Error("expected at least one step")
	}
}

func TestDiscover_GitInstallHooksIsResolvable(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	a, err := result.Find("git/install-hooks")
	if err != nil {
		t.Fatalf("Find('git/install-hooks') returned error: %v", err)
	}
	if a.Name != "git/install-hooks" {
		t.Errorf("expected name 'git/install-hooks', got %q", a.Name)
	}
}

func TestDiscover_GitInstallHooksUsesBash(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	a := result.Automations["git/install-hooks"]
	if len(a.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(a.Steps))
	}
	if a.Steps[0].Type != automation.StepTypeBash {
		t.Errorf("expected bash step, got %q", a.Steps[0].Type)
	}
}

func TestDiscover_GitInstallHooksAcceptsSourceInput(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	a := result.Automations["git/install-hooks"]
	if len(a.Inputs) == 0 {
		t.Fatal("expected git/install-hooks to have inputs")
	}
	spec, ok := a.Inputs["source"]
	if !ok {
		t.Fatal("expected git/install-hooks to have a 'source' input")
	}
	if !spec.IsRequired() {
		t.Error("expected 'source' input to be required")
	}
	if spec.Description == "" {
		t.Error("expected 'source' input to have a description")
	}
}

func TestDiscover_GitInstallHooksUsesInput(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	a := result.Automations["git/install-hooks"]
	script := a.Steps[0].Value
	if !strings.Contains(script, "PI_INPUT_SOURCE") {
		t.Error("expected script to use PI_INPUT_SOURCE env var")
	}
}

func TestDiscover_GitInstallHooksIsIdempotent(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	a := result.Automations["git/install-hooks"]
	script := a.Steps[0].Value
	if !strings.Contains(script, "cmp") {
		t.Error("expected script to compare files for idempotency (cmp)")
	}
	if !strings.Contains(script, "[already installed]") {
		t.Error("expected script to print '[already installed]' when hooks are up to date")
	}
	if !strings.Contains(script, "[installed]") {
		t.Error("expected script to print '[installed]' after installing hooks")
	}
}

func TestDiscover_GitInstallHooksMakesExecutable(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	a := result.Automations["git/install-hooks"]
	script := a.Steps[0].Value
	if !strings.Contains(script, "chmod +x") {
		t.Error("expected script to make hooks executable via chmod +x")
	}
}

func TestDiscover_GitInstallHooksCopiesToGitHooksDir(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	a := result.Automations["git/install-hooks"]
	script := a.Steps[0].Value
	if !strings.Contains(script, ".git/hooks") {
		t.Error("expected script to reference .git/hooks directory")
	}
}

func TestDiscover_DevToolAutomationsNoInputs(t *testing.T) {
	result, err := Discover()
	if err != nil {
		t.Fatalf("Discover() returned error: %v", err)
	}

	// cursor/install-extensions requires 'extensions' input
	a := result.Automations["cursor/install-extensions"]
	if len(a.Inputs) != 1 {
		t.Errorf("expected cursor/install-extensions to have 1 input, got %d", len(a.Inputs))
	}

	// git/install-hooks requires 'source' input
	a = result.Automations["git/install-hooks"]
	if len(a.Inputs) != 1 {
		t.Errorf("expected git/install-hooks to have 1 input, got %d", len(a.Inputs))
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
