package integration

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestBuiltins_List_ShowsBuiltinMarker(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "[built-in]") {
		t.Errorf("expected [built-in] marker in list output, got:\n%s", out)
	}
}

func TestBuiltins_RunWithPiPrefix(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	out, code := runPi(t, dir, "run", "pi:hello")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "hello from built-in") {
		t.Errorf("expected built-in output, got:\n%s", out)
	}
}

func TestBuiltins_LocalShadowsBuiltin(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	out, code := runPi(t, dir, "run", "hello")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "hello from local override") {
		t.Errorf("expected local override, got:\n%s", out)
	}
}

func TestBuiltins_PiPrefixAlwaysGetsBuiltin(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	out, code := runPi(t, dir, "run", "pi:hello")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "hello from built-in") {
		t.Errorf("expected built-in despite local shadow, got:\n%s", out)
	}
}

func TestBuiltins_RunStepCallsBuiltin(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	out, code := runPi(t, dir, "run", "call-builtin")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "hello from built-in") {
		t.Errorf("expected run step to resolve pi:hello to built-in, got:\n%s", out)
	}
}

func TestBuiltins_InfoWithPiPrefix(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	out, code := runPi(t, dir, "info", "pi:hello")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "Name:") {
		t.Errorf("expected info output, got:\n%s", out)
	}
	if !strings.Contains(out, "hello") {
		t.Errorf("expected automation name in info, got:\n%s", out)
	}
}

func TestBuiltins_ListShadowed(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "hello") && strings.Contains(line, "[built-in]") {
			t.Errorf("expected local 'hello' to NOT have [built-in] marker, got:\n%s", line)
		}
	}
}

func TestBuiltins_SetupWithPiPrefix(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	out, code := runPiWithEnv(t, dir, []string{"CI=true"}, "setup")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "hello from built-in") {
		t.Errorf("expected setup to run pi:hello built-in, got:\n%s", out)
	}
	if !strings.Contains(out, "hello from local") {
		t.Errorf("expected setup to run local-hello, got:\n%s", out)
	}
}

func TestBuiltins_PiPrefixNotFound(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	out, code := runPi(t, dir, "run", "pi:nonexistent")
	if code == 0 {
		t.Fatalf("expected non-zero exit for pi:nonexistent")
	}
	if !strings.Contains(out, "built-in automation") {
		t.Errorf("expected built-in not found error, got:\n%s", out)
	}
}

func TestBuiltins_DockerAutomationsInList(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	for _, name := range []string{"docker/up", "docker/down", "docker/logs"} {
		if !strings.Contains(out, name) {
			t.Errorf("expected %q in list output, got:\n%s", name, out)
		}
	}
}

func TestBuiltins_DockerAutomationsMarkedBuiltIn(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	for _, name := range []string{"docker/up", "docker/down", "docker/logs"} {
		found := false
		for _, line := range strings.Split(out, "\n") {
			if strings.Contains(line, name) && strings.Contains(line, "[built-in]") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected %q to have [built-in] marker in list output, got:\n%s", name, out)
		}
	}
}

func TestBuiltins_DockerInfoShowsDetails(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	for _, name := range []string{"pi:docker/up", "pi:docker/down", "pi:docker/logs"} {
		t.Run(name, func(t *testing.T) {
			out, code := runPi(t, dir, "info", name)
			if code != 0 {
				t.Fatalf("expected exit 0, got %d: %s", code, out)
			}
			if !strings.Contains(out, "Name:") {
				t.Errorf("expected Name: in info output, got:\n%s", out)
			}
			if !strings.Contains(out, "Description:") {
				t.Errorf("expected Description: in info output, got:\n%s", out)
			}
		})
	}
}

func TestBuiltins_DockerRunStepResolvesBuiltin(t *testing.T) {
	dir := filepath.Join(examplesDir(), "docker-builtins")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "call-docker-up") {
		t.Errorf("expected call-docker-up in list output, got:\n%s", out)
	}
	if !strings.Contains(out, "docker/up") {
		t.Errorf("expected docker/up built-in in list output, got:\n%s", out)
	}
}

func TestBuiltins_DockerRunStepInfoResolvesBuiltin(t *testing.T) {
	dir := filepath.Join(examplesDir(), "docker-builtins")
	out, code := runPi(t, dir, "info", "pi:docker/up")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "docker/up") {
		t.Errorf("expected docker/up in info output, got:\n%s", out)
	}
	if !strings.Contains(out, "Docker Compose") {
		t.Errorf("expected description mentioning Docker Compose, got:\n%s", out)
	}
}

func TestBuiltins_InstallerAutomationsMarkedBuiltIn(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	for _, name := range []string{"install-homebrew", "install-python", "install-node", "install-go", "install-rust", "install-uv", "install-tsx"} {
		found := false
		for _, line := range strings.Split(out, "\n") {
			if strings.Contains(line, name) && strings.Contains(line, "[built-in]") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected %q to have [built-in] marker in list output, got:\n%s", name, out)
		}
	}
}

func TestBuiltins_InstallerInfoShowsDetails(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	installers := []struct {
		name   string
		substr string
	}{
		{"pi:install-homebrew", "Homebrew"},
		{"pi:install-python", "Python"},
		{"pi:install-node", "Node.js"},
		{"pi:install-rust", "Rust"},
		{"pi:install-uv", "uv"},
		{"pi:install-tsx", "tsx"},
	}
	for _, tc := range installers {
		t.Run(tc.name, func(t *testing.T) {
			out, code := runPi(t, dir, "info", tc.name)
			if code != 0 {
				t.Fatalf("expected exit 0, got %d: %s", code, out)
			}
			if !strings.Contains(out, "Name:") {
				t.Errorf("expected Name: in info output, got:\n%s", out)
			}
			if !strings.Contains(out, tc.substr) {
				t.Errorf("expected %q in info output, got:\n%s", tc.substr, out)
			}
		})
	}
}

func TestBuiltins_InstallerInfoShowsInputs(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	for _, name := range []string{"pi:install-python", "pi:install-node", "pi:install-rust"} {
		t.Run(name, func(t *testing.T) {
			out, code := runPi(t, dir, "info", name)
			if code != 0 {
				t.Fatalf("expected exit 0, got %d: %s", code, out)
			}
			if !strings.Contains(out, "version") {
				t.Errorf("expected 'version' input in info output, got:\n%s", out)
			}
		})
	}
}

func TestBuiltins_InstallerHomebrewShowsCondition(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	out, code := runPi(t, dir, "info", "pi:install-homebrew")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "os.macos") {
		t.Errorf("expected 'os.macos' condition in info output, got:\n%s", out)
	}
}

func TestBuiltins_InstallTsxIdempotent(t *testing.T) {
	requireTsx(t)
	dir := filepath.Join(examplesDir(), "builtins")
	out, code := runPi(t, dir, "run", "pi:install-tsx")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "already installed") && !strings.Contains(out, "installed") {
		t.Errorf("expected 'already installed' or 'installed' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "✓") {
		t.Errorf("expected '✓' status icon in output, got:\n%s", out)
	}
}

func TestBuiltins_InstallerListShowsInputsColumn(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "install-python") && strings.Contains(line, "[built-in]") {
			if !strings.Contains(line, "version") {
				t.Errorf("expected install-python list line to show 'version' input, got:\n%s", line)
			}
			break
		}
	}
}

func TestBuiltins_DevToolAutomationsInList(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	for _, name := range []string{"cursor/install-extensions", "git/install-hooks"} {
		if !strings.Contains(out, name) {
			t.Errorf("expected %q in list output, got:\n%s", name, out)
		}
	}
}

func TestBuiltins_DevToolAutomationsMarkedBuiltIn(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	for _, name := range []string{"cursor/install-extensions", "git/install-hooks"} {
		found := false
		for _, line := range strings.Split(out, "\n") {
			if strings.Contains(line, name) && strings.Contains(line, "[built-in]") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected %q to have [built-in] marker in list output, got:\n%s", name, out)
		}
	}
}

func TestBuiltins_DevToolInfoShowsDetails(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	tests := []struct {
		name        string
		description string
	}{
		{"pi:cursor/install-extensions", "Install missing Cursor extensions"},
		{"pi:git/install-hooks", "Install git hooks from a source directory"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, code := runPi(t, dir, "info", tc.name)
			if code != 0 {
				t.Fatalf("expected exit 0, got %d: %s", code, out)
			}
			if !strings.Contains(out, "Name:") {
				t.Errorf("expected Name: in info output, got:\n%s", out)
			}
			if !strings.Contains(out, "Description:") {
				t.Errorf("expected Description: in info output, got:\n%s", out)
			}
		})
	}
}

func TestBuiltins_DevToolInfoShowsInputs(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	tests := []struct {
		name  string
		input string
	}{
		{"pi:cursor/install-extensions", "extensions"},
		{"pi:git/install-hooks", "source"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, code := runPi(t, dir, "info", tc.name)
			if code != 0 {
				t.Fatalf("expected exit 0, got %d: %s", code, out)
			}
			if !strings.Contains(out, tc.input) {
				t.Errorf("expected %q input in info output, got:\n%s", tc.input, out)
			}
		})
	}
}

func TestBuiltins_DevToolListShowsInputsColumn(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}

	tests := []struct {
		name  string
		input string
	}{
		{"cursor/install-extensions", "extensions"},
		{"git/install-hooks", "source"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for _, line := range strings.Split(out, "\n") {
				if strings.Contains(line, tc.name) && strings.Contains(line, "[built-in]") {
					if !strings.Contains(line, tc.input) {
						t.Errorf("expected %s list line to show %q input, got:\n%s", tc.name, tc.input, line)
					}
					return
				}
			}
			t.Errorf("did not find %s in list output:\n%s", tc.name, out)
		})
	}
}
