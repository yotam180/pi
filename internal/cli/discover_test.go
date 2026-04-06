package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vyper-tooling/pi/internal/config"
	"github.com/vyper-tooling/pi/internal/display"
)

func TestPrintOnDemandAdvisory_OutputFormat(t *testing.T) {
	var buf bytes.Buffer
	printOnDemandAdvisory(&buf, "yotam180/pi-common@v1.2")

	output := buf.String()

	if !strings.Contains(output, "yotam180/pi-common@v1.2") {
		t.Errorf("expected source in output, got: %q", output)
	}
	if !strings.Contains(output, "tip:") {
		t.Errorf("expected 'tip:' in advisory, got: %q", output)
	}
	if !strings.Contains(output, "packages:") {
		t.Errorf("expected 'packages:' snippet in advisory, got: %q", output)
	}
	if !strings.Contains(output, "- yotam180/pi-common@v1.2") {
		t.Errorf("expected ready-to-paste snippet, got: %q", output)
	}
}

func TestPrintOnDemandAdvisory_NilWriter(t *testing.T) {
	// Should not panic
	printOnDemandAdvisory(nil, "org/repo@v1.0")
}

func TestPrintOnDemandAdvisory_ContainsFetchStatus(t *testing.T) {
	var buf bytes.Buffer
	printOnDemandAdvisory(&buf, "org/repo@v2.0")

	output := buf.String()

	if !strings.Contains(output, "fetched (on demand)") {
		t.Errorf("expected 'fetched (on demand)' status, got: %q", output)
	}
}

func TestPrintOnDemandAdvisory_ContainsDownArrow(t *testing.T) {
	var buf bytes.Buffer
	printOnDemandAdvisory(&buf, "org/repo@v1.0")

	output := buf.String()

	if !strings.Contains(output, "↓") {
		t.Errorf("expected down arrow icon, got: %q", output)
	}
}

func TestResolveFilePackage_ExistingDir(t *testing.T) {
	root := t.TempDir()
	pkgDir := filepath.Join(root, "my-pkg")
	os.MkdirAll(filepath.Join(pkgDir, ".pi"), 0755)

	var buf bytes.Buffer
	printer := display.NewForWriter(&buf)
	pkg := config.PackageEntry{Source: "file:" + pkgDir}

	got, err := resolveFilePackage(pkg, root, &buf, printer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != pkgDir {
		t.Errorf("got %q, want %q", got, pkgDir)
	}
	if !strings.Contains(buf.String(), "found") {
		t.Errorf("expected 'found' status in output, got: %q", buf.String())
	}
}

func TestResolveFilePackage_ExistingDirWithAlias(t *testing.T) {
	root := t.TempDir()
	pkgDir := filepath.Join(root, "my-pkg")
	os.MkdirAll(pkgDir, 0755)

	var buf bytes.Buffer
	printer := display.NewForWriter(&buf)
	pkg := config.PackageEntry{Source: "file:" + pkgDir, As: "mytools"}

	got, err := resolveFilePackage(pkg, root, &buf, printer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != pkgDir {
		t.Errorf("got %q, want %q", got, pkgDir)
	}
	if !strings.Contains(buf.String(), "alias: mytools") {
		t.Errorf("expected alias detail in output, got: %q", buf.String())
	}
}

func TestResolveFilePackage_MissingDir(t *testing.T) {
	root := t.TempDir()

	var buf bytes.Buffer
	printer := display.NewForWriter(&buf)
	pkg := config.PackageEntry{Source: "file:" + filepath.Join(root, "nonexistent")}

	got, err := resolveFilePackage(pkg, root, &buf, printer)
	if err != nil {
		t.Fatalf("unexpected error (should be non-fatal): %v", err)
	}
	if got != "" {
		t.Errorf("expected empty path for missing dir, got %q", got)
	}
	if !strings.Contains(buf.String(), "not found") {
		t.Errorf("expected 'not found' status, got: %q", buf.String())
	}
	if !strings.Contains(buf.String(), "warning:") {
		t.Errorf("expected warning message, got: %q", buf.String())
	}
}

func TestResolveFilePackage_MissingDirWithAlias(t *testing.T) {
	root := t.TempDir()

	var buf bytes.Buffer
	printer := display.NewForWriter(&buf)
	pkg := config.PackageEntry{Source: "file:" + filepath.Join(root, "gone"), As: "tools"}

	got, err := resolveFilePackage(pkg, root, &buf, printer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty path, got %q", got)
	}
	if !strings.Contains(buf.String(), "alias: tools") {
		t.Errorf("expected alias in not-found output, got: %q", buf.String())
	}
}

func TestResolveFilePackage_NotADir(t *testing.T) {
	root := t.TempDir()
	filePath := filepath.Join(root, "not-a-dir")
	os.WriteFile(filePath, []byte("i am a file"), 0644)

	var buf bytes.Buffer
	printer := display.NewForWriter(&buf)
	pkg := config.PackageEntry{Source: "file:" + filePath}

	got, err := resolveFilePackage(pkg, root, &buf, printer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty path for non-directory, got %q", got)
	}
	if !strings.Contains(buf.String(), "not found") {
		t.Errorf("expected 'not found' status for non-directory, got: %q", buf.String())
	}
}

func TestResolveFilePackage_RelativePath(t *testing.T) {
	root := t.TempDir()
	pkgDir := filepath.Join(root, "packages", "shared")
	os.MkdirAll(pkgDir, 0755)

	var buf bytes.Buffer
	printer := display.NewForWriter(&buf)
	pkg := config.PackageEntry{Source: "file:./packages/shared"}

	got, err := resolveFilePackage(pkg, root, &buf, printer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != pkgDir {
		t.Errorf("relative path should resolve to %q, got %q", pkgDir, got)
	}
}

func TestResolveFilePackage_NilPrinter(t *testing.T) {
	root := t.TempDir()
	pkgDir := filepath.Join(root, "pkg")
	os.MkdirAll(pkgDir, 0755)

	var buf bytes.Buffer
	pkg := config.PackageEntry{Source: "file:" + pkgDir}

	got, err := resolveFilePackage(pkg, root, &buf, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != pkgDir {
		t.Errorf("got %q, want %q", got, pkgDir)
	}
}

func TestResolveFilePackage_NilStderr(t *testing.T) {
	root := t.TempDir()
	pkg := config.PackageEntry{Source: "file:" + filepath.Join(root, "gone")}

	got, err := resolveFilePackage(pkg, root, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty path, got %q", got)
	}
}

func TestResolvePackageSource_FileRouting(t *testing.T) {
	root := t.TempDir()
	pkgDir := filepath.Join(root, "my-pkg")
	os.MkdirAll(pkgDir, 0755)

	pkg := config.PackageEntry{Source: "file:" + pkgDir}
	got, err := resolvePackageSource(pkg, root, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != pkgDir {
		t.Errorf("file source should resolve to %q, got %q", pkgDir, got)
	}
}

func TestMergePackages_EmptyList(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(root, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "hello.yaml"), []byte("description: hello\nbash: echo hi\n"), 0644)

	cfg := &config.ProjectConfig{
		Project:  "test",
		Packages: []config.PackageEntry{},
	}

	result, err := discoverAllWithConfig(root, cfg, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.Automations["hello"]; !ok {
		t.Error("expected hello automation to be discovered")
	}
}

func TestMergePackages_FileSourceSkippedWhenMissing(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(root, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "hello.yaml"), []byte("description: hello\nbash: echo hi\n"), 0644)

	cfg := &config.ProjectConfig{
		Project: "test",
		Packages: []config.PackageEntry{
			{Source: "file:" + filepath.Join(root, "nonexistent")},
		},
	}

	var buf bytes.Buffer
	result, err := discoverAllWithConfig(root, cfg, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.Automations["hello"]; !ok {
		t.Error("local automations should still be discovered when file: source is missing")
	}
	if !strings.Contains(buf.String(), "warning:") {
		t.Errorf("expected warning about missing path, got: %q", buf.String())
	}
}
