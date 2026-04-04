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
