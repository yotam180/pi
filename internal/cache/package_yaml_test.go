package cache

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckPackageYAML_Absent(t *testing.T) {
	dir := t.TempDir()
	c := &Cache{PIVersion: "1.0.0"}
	if err := c.checkPackageYAML(dir); err != nil {
		t.Errorf("expected nil error for absent pi-package.yaml, got: %v", err)
	}
}

func TestCheckPackageYAML_EmptyMinVersion(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi-package.yaml"), []byte("min_pi_version: \"\""), 0o644)

	c := &Cache{PIVersion: "1.0.0"}
	if err := c.checkPackageYAML(dir); err != nil {
		t.Errorf("expected nil error for empty min_pi_version, got: %v", err)
	}
}

func TestCheckPackageYAML_NoMinVersionField(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi-package.yaml"), []byte("name: my-pkg"), 0o644)

	c := &Cache{PIVersion: "1.0.0"}
	if err := c.checkPackageYAML(dir); err != nil {
		t.Errorf("expected nil error for missing min_pi_version field, got: %v", err)
	}
}

func TestCheckPackageYAML_Satisfied(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi-package.yaml"), []byte("min_pi_version: \"1.0\""), 0o644)

	c := &Cache{PIVersion: "1.2.3"}
	if err := c.checkPackageYAML(dir); err != nil {
		t.Errorf("expected nil error for satisfied version, got: %v", err)
	}
}

func TestCheckPackageYAML_ExactMatch(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi-package.yaml"), []byte("min_pi_version: \"1.2.3\""), 0o644)

	c := &Cache{PIVersion: "1.2.3"}
	if err := c.checkPackageYAML(dir); err != nil {
		t.Errorf("expected nil error for exact version match, got: %v", err)
	}
}

func TestCheckPackageYAML_Unsatisfied(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi-package.yaml"), []byte("min_pi_version: \"2.0\""), 0o644)

	c := &Cache{PIVersion: "1.5.0"}
	err := c.checkPackageYAML(dir)
	if err == nil {
		t.Fatal("expected error for unsatisfied min_pi_version")
	}
	if got := err.Error(); !strings.Contains(got, "requires PI >= 2.0") || !strings.Contains(got, "running PI 1.5.0") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestCheckPackageYAML_DevBuild_SkipsCheck(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi-package.yaml"), []byte("min_pi_version: \"99.0\""), 0o644)

	c := &Cache{PIVersion: "dev"}
	if err := c.checkPackageYAML(dir); err != nil {
		t.Errorf("dev builds should skip version check, got: %v", err)
	}
}

func TestCheckPackageYAML_EmptyPIVersion_SkipsCheck(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi-package.yaml"), []byte("min_pi_version: \"99.0\""), 0o644)

	c := &Cache{PIVersion: ""}
	if err := c.checkPackageYAML(dir); err != nil {
		t.Errorf("empty PIVersion should skip version check, got: %v", err)
	}
}

func TestCheckPackageYAML_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi-package.yaml"), []byte("{{invalid yaml"), 0o644)

	c := &Cache{PIVersion: "1.0.0"}
	err := c.checkPackageYAML(dir)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
	if !strings.Contains(err.Error(), "parsing") {
		t.Errorf("error should mention parsing: %v", err)
	}
}

func TestCheckPackageYAML_WithVPrefix(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi-package.yaml"), []byte("min_pi_version: v1.0"), 0o644)

	c := &Cache{PIVersion: "v1.5.0"}
	if err := c.checkPackageYAML(dir); err != nil {
		t.Errorf("expected nil error with v-prefix versions, got: %v", err)
	}
}

func TestVersionSatisfies(t *testing.T) {
	tests := []struct {
		running  string
		required string
		want     bool
	}{
		{"1.0.0", "1.0.0", true},
		{"1.1.0", "1.0.0", true},
		{"1.0.1", "1.0.0", true},
		{"2.0.0", "1.0.0", true},
		{"0.9.0", "1.0.0", false},
		{"1.0.0", "1.0.1", false},
		{"1.0", "1.0.0", true},   // missing component treated as 0
		{"1.0.0", "1.0", true},   // missing component treated as 0
		{"v1.0.0", "1.0.0", true}, // v-prefix stripped
		{"1.0.0", "v1.0.0", true}, // v-prefix stripped
		{"1.5", "1.2", true},
		{"1.2", "1.5", false},
	}

	for _, tt := range tests {
		got := versionSatisfies(tt.running, tt.required)
		if got != tt.want {
			t.Errorf("versionSatisfies(%q, %q) = %v, want %v",
				tt.running, tt.required, got, tt.want)
		}
	}
}

func TestFetch_PackageYAML_Satisfied(t *testing.T) {
	cacheRoot := t.TempDir()

	fg := newFakeGit()
	fg.allowedURLs["https://github.com/myorg/myrepo.git"] = true
	fg.allowedVersions["v1.0"] = true
	fg.repoFiles["pi-package.yaml"] = "min_pi_version: \"0.5\""

	c := &Cache{
		Root:       cacheRoot,
		PIVersion:  "1.0.0",
		GitFunc:    fg.gitFunc,
		GetenvFunc: func(string) string { return "" },
	}

	path, err := c.Fetch("myorg", "myrepo", "v1.0")
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}
	if path == "" {
		t.Fatal("expected non-empty path")
	}
}

func TestFetch_PackageYAML_Unsatisfied_CleansUp(t *testing.T) {
	cacheRoot := t.TempDir()

	fg := newFakeGit()
	fg.allowedURLs["https://github.com/myorg/myrepo.git"] = true
	fg.allowedVersions["v1.0"] = true
	fg.repoFiles["pi-package.yaml"] = "min_pi_version: \"99.0\""

	c := &Cache{
		Root:       cacheRoot,
		PIVersion:  "1.0.0",
		GitFunc:    fg.gitFunc,
		GetenvFunc: func(string) string { return "" },
	}

	_, err := c.Fetch("myorg", "myrepo", "v1.0")
	if err == nil {
		t.Fatal("expected error for unsatisfied min_pi_version")
	}

	// Cache entry should be cleaned up
	target := c.PackagePath("myorg", "myrepo", "v1.0")
	if _, statErr := os.Stat(target); !os.IsNotExist(statErr) {
		t.Errorf("expected cache entry to be cleaned up at %s", target)
	}
}

func TestFetch_PackageYAML_Absent_Proceeds(t *testing.T) {
	cacheRoot := t.TempDir()

	fg := newFakeGit()
	fg.allowedURLs["https://github.com/myorg/myrepo.git"] = true
	fg.allowedVersions["v1.0"] = true
	// No pi-package.yaml in repoFiles

	c := &Cache{
		Root:       cacheRoot,
		PIVersion:  "1.0.0",
		GitFunc:    fg.gitFunc,
		GetenvFunc: func(string) string { return "" },
	}

	path, err := c.Fetch("myorg", "myrepo", "v1.0")
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}
	if path == "" {
		t.Fatal("expected non-empty path")
	}
}

func TestFetch_CacheHit_PackageYAML_Checked(t *testing.T) {
	cacheRoot := t.TempDir()
	target := filepath.Join(cacheRoot, "github", "myorg", "myrepo", "v1.0")
	os.MkdirAll(target, 0o755)
	os.WriteFile(filepath.Join(target, "pi-package.yaml"), []byte("min_pi_version: \"99.0\""), 0o644)

	c := &Cache{
		Root:      cacheRoot,
		PIVersion: "1.0.0",
	}

	_, err := c.Fetch("myorg", "myrepo", "v1.0")
	if err == nil {
		t.Fatal("expected error even on cache hit when min_pi_version is unsatisfied")
	}
}

