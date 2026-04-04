package builtins

import (
	"testing"
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
