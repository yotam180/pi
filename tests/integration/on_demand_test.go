package integration

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestOnDemand_FileRef_NeverOnDemand(t *testing.T) {
	dir := t.TempDir()

	writeIntegFile(t, filepath.Join(dir, "pi.yaml"), `
project: on-demand-test
`)

	writeIntegFile(t, filepath.Join(dir, ".pi", "caller.yaml"), `
description: Calls a file ref
steps:
  - run: "file:/tmp/nonexistent-pi-package-12345"
`)

	_, stderr, code := runPiSplit(t, dir, "run", "caller")
	if code == 0 {
		t.Fatal("expected non-zero exit for missing file: ref")
	}

	combined := stderr
	if !strings.Contains(combined, "not found") {
		t.Errorf("expected 'not found' error for file: ref, got:\n%s", combined)
	}
	if strings.Contains(combined, "on demand") {
		t.Error("file: refs should never trigger on-demand fetch")
	}
}

func TestOnDemand_GitHubRef_Undeclared_ShowsAdvisory(t *testing.T) {
	dir := t.TempDir()

	writeIntegFile(t, filepath.Join(dir, "pi.yaml"), `
project: on-demand-test
`)

	writeIntegFile(t, filepath.Join(dir, ".pi", "caller.yaml"), `
description: Calls an undeclared GitHub ref
steps:
  - run: nonexistent-org-xyz/nonexistent-repo-abc@v0.0.1/some/automation
`)

	// This will fail because the repo doesn't exist, but the error
	// should come from the on-demand fetch attempt (not a "not supported" error)
	_, stderr, code := runPiSplit(t, dir, "run", "caller")
	if code == 0 {
		t.Fatal("expected non-zero exit for unfetchable GitHub ref")
	}

	if !strings.Contains(stderr, "on-demand fetch") {
		t.Errorf("expected 'on-demand fetch' in error for GitHub ref, got stderr:\n%s", stderr)
	}
}

func TestOnDemand_DeclaredPackage_NoAdvisory(t *testing.T) {
	dir := t.TempDir()

	// Create a local file: package with an automation
	writeIntegFile(t, filepath.Join(dir, "pkg", ".pi", "hello.yaml"), `
description: Hello from package
bash: echo "package hello"
`)

	writeIntegFile(t, filepath.Join(dir, "pi.yaml"), `
project: on-demand-test
packages:
  - source: file:./pkg
    as: mypkg
`)

	writeIntegFile(t, filepath.Join(dir, ".pi", "dummy.yaml"), `
description: Dummy
bash: echo dummy
`)

	// Running a package automation by plain name should work silently
	stdout, stderr, code := runPiSplit(t, dir, "run", "hello")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\nstdout: %s\nstderr: %s", code, stdout, stderr)
	}
	if !strings.Contains(stdout, "package hello") {
		t.Errorf("expected 'package hello' output, got:\n%s", stdout)
	}
	if strings.Contains(stderr, "tip:") {
		t.Error("declared packages should NOT show advisory tip")
	}
}

func TestOnDemand_CachedPackage_Silent(t *testing.T) {
	dir := t.TempDir()

	// Simulate a "cached" package by creating a file: package
	// (file: packages are always "found" on disk — no advisory)
	writeIntegFile(t, filepath.Join(dir, "pkg", ".pi", "deploy.yaml"), `
description: Deploy
bash: echo deployed
`)

	writeIntegFile(t, filepath.Join(dir, "pi.yaml"), `
project: cache-test
packages:
  - source: file:./pkg
`)

	writeIntegFile(t, filepath.Join(dir, ".pi", "local.yaml"), `
description: Local
bash: echo local
`)

	stdout, stderr, code := runPiSplit(t, dir, "run", "deploy")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\nstdout: %s\nstderr: %s", code, stdout, stderr)
	}
	if strings.Contains(stderr, "on demand") {
		t.Error("cached/declared packages should not show on-demand advisory")
	}
	if strings.Contains(stderr, "tip:") {
		t.Error("cached/declared packages should not show tip")
	}
}

func TestOnDemand_FileRef_MissingPath_ClearError(t *testing.T) {
	dir := t.TempDir()

	writeIntegFile(t, filepath.Join(dir, "pi.yaml"), `
project: file-error-test
packages:
  - source: file:./nonexistent-pkg
`)

	writeIntegFile(t, filepath.Join(dir, ".pi", "local.yaml"), `
description: Local
bash: echo local
`)

	// pi setup passes stderr, so warnings about missing file: packages are visible
	stdout, stderr, code := runPiSplit(t, dir, "setup", "--no-shell")
	_ = code // setup may or may not fail depending on config
	_ = stdout

	if !strings.Contains(stderr, "not found") || !strings.Contains(stderr, "nonexistent-pkg") {
		t.Errorf("expected warning about missing file: package in stderr, got:\n%s", stderr)
	}
}

func TestOnDemand_AdvisoryToStderr(t *testing.T) {
	dir := t.TempDir()

	writeIntegFile(t, filepath.Join(dir, "pi.yaml"), `
project: stderr-test
`)

	writeIntegFile(t, filepath.Join(dir, ".pi", "caller.yaml"), `
description: Calls undeclared GitHub ref
steps:
  - run: nonexistent-org-12345/nonexistent-repo-67890@v0.0.1/hello
`)

	// The fetch will fail, but let's verify the error goes to stderr not stdout
	stdout, stderr, code := runPiSplit(t, dir, "run", "caller")
	if code == 0 {
		t.Fatal("expected non-zero exit for unfetchable GitHub ref")
	}
	_ = stdout

	// Error output should be on stderr
	if !strings.Contains(stderr, "nonexistent-org-12345/nonexistent-repo-67890@v0.0.1") {
		t.Errorf("expected package reference in stderr, got:\n%s", stderr)
	}
}

