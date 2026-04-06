package executor

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// These are smoke tests verifying the executor's delegation to conditions package.
// Comprehensive predicate tests live in internal/conditions/evaluator_test.go.

func TestDelegatedResolvePredicates(t *testing.T) {
	env := &RuntimeEnv{
		GOOS:   "darwin",
		GOARCH: "arm64",
		Getenv: func(key string) string {
			if key == "SHELL" {
				return "/bin/zsh"
			}
			return ""
		},
		LookPath: func(name string) (string, error) {
			if name == "docker" {
				return "/usr/bin/docker", nil
			}
			return "", fmt.Errorf("not found")
		},
		Stat: os.Stat,
	}

	preds := []string{"os.macos", "os.arch.arm64", "shell.zsh", "command.docker"}
	result, err := ResolvePredicatesWithEnv(preds, "/tmp", env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, p := range preds {
		if !result[p] {
			t.Errorf("expected %q to be true", p)
		}
	}
}

func TestDelegatedDefaultRuntimeEnv(t *testing.T) {
	env := DefaultRuntimeEnv()
	if env.GOOS == "" {
		t.Error("GOOS should not be empty")
	}
	if env.GOARCH == "" {
		t.Error("GOARCH should not be empty")
	}
}

func TestDelegatedResolvePredicatesIntegration(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "test.txt"), []byte("hi"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := ResolvePredicates(
		[]string{`file.exists("test.txt")`}, dir,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result[`file.exists("test.txt")`] {
		t.Error("expected true from delegated ResolvePredicates")
	}
}

func TestDelegatedValidatePredicateName(t *testing.T) {
	if err := ValidatePredicateName("os.macos"); err != nil {
		t.Errorf("expected valid: %v", err)
	}
	if err := ValidatePredicateName("os.macoss"); err == nil {
		t.Error("expected error for typo")
	}
}

func TestDelegatedValidateConditionExpr(t *testing.T) {
	if err := ValidateConditionExpr("os.macos and command.docker"); err != nil {
		t.Errorf("expected valid: %v", err)
	}
	if err := ValidateConditionExpr("bogus.thing"); err == nil {
		t.Error("expected error for unknown predicate")
	}
}
