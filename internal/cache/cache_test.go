package cache

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeGit simulates git clone + checkout behavior for tests.
type fakeGit struct {
	// cloneURLs records which URLs were attempted.
	cloneURLs []string
	// allowedURLs is the set of URLs that succeed.
	allowedURLs map[string]bool
	// allowedVersions is the set of versions that checkout succeeds for.
	allowedVersions map[string]bool
	// repoFiles maps filename → content to write into the cloned dir.
	repoFiles map[string]string
}

func newFakeGit() *fakeGit {
	return &fakeGit{
		allowedURLs:     make(map[string]bool),
		allowedVersions: make(map[string]bool),
		repoFiles:       make(map[string]string),
	}
}

func (f *fakeGit) gitFunc(args []string, dir string) (string, string, error) {
	if len(args) == 0 {
		return "", "", fmt.Errorf("no args")
	}

	switch args[0] {
	case "clone":
		return f.handleClone(args, dir)
	case "checkout":
		return f.handleCheckout(args, dir)
	default:
		return "", "", fmt.Errorf("unsupported git command: %s", args[0])
	}
}

func (f *fakeGit) handleClone(args []string, _ string) (string, string, error) {
	// Find the URL (skip flags)
	var url, dest string
	for i := 1; i < len(args); i++ {
		if strings.HasPrefix(args[i], "--") {
			continue
		}
		if url == "" {
			url = args[i]
		} else {
			dest = args[i]
		}
	}

	f.cloneURLs = append(f.cloneURLs, url)

	if !f.allowedURLs[url] {
		return "", "fatal: repository not found", fmt.Errorf("clone failed")
	}

	// Create .git dir to simulate a real clone
	if dest != "" {
		os.MkdirAll(filepath.Join(dest, ".git"), 0o755)
		for name, content := range f.repoFiles {
			p := filepath.Join(dest, name)
			os.MkdirAll(filepath.Dir(p), 0o755)
			os.WriteFile(p, []byte(content), 0o644)
		}
	}

	return "", "", nil
}

func (f *fakeGit) handleCheckout(args []string, dir string) (string, string, error) {
	var version string
	for i := 1; i < len(args); i++ {
		if !strings.HasPrefix(args[i], "--") {
			version = args[i]
			break
		}
	}

	if !f.allowedVersions[version] {
		return "", fmt.Sprintf("error: pathspec '%s' did not match any file(s) known to git", version),
			fmt.Errorf("checkout failed")
	}

	return "", "", nil
}

func TestDefaultCacheRoot(t *testing.T) {
	root, err := DefaultCacheRoot()
	if err != nil {
		t.Fatalf("DefaultCacheRoot() error: %v", err)
	}
	if !strings.HasSuffix(root, filepath.Join(".pi", "cache")) {
		t.Errorf("expected root to end with .pi/cache, got %s", root)
	}
}

func TestPackagePath(t *testing.T) {
	c := &Cache{Root: "/home/user/.pi/cache"}
	got := c.PackagePath("yotam180", "pi-common", "v1.2")
	want := filepath.Join("/home/user/.pi/cache", "github", "yotam180", "pi-common", "v1.2")
	if got != want {
		t.Errorf("PackagePath = %q, want %q", got, want)
	}
}

func TestIsMutableRef(t *testing.T) {
	tests := []struct {
		version string
		want    bool
	}{
		{"main", true},
		{"master", true},
		{"HEAD", true},
		{"v1.0", false},
		{"v2.3.4", false},
		{"feature-branch", false},
	}
	for _, tt := range tests {
		if got := IsMutableRef(tt.version); got != tt.want {
			t.Errorf("IsMutableRef(%q) = %v, want %v", tt.version, got, tt.want)
		}
	}
}

func TestFetch_CacheHit(t *testing.T) {
	cacheRoot := t.TempDir()
	target := filepath.Join(cacheRoot, "github", "myorg", "myrepo", "v1.0")
	os.MkdirAll(target, 0o755)

	fg := newFakeGit()
	c := &Cache{Root: cacheRoot, GitFunc: fg.gitFunc}

	path, err := c.Fetch("myorg", "myrepo", "v1.0")
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}
	if path != target {
		t.Errorf("Fetch() = %q, want %q", path, target)
	}
	if len(fg.cloneURLs) != 0 {
		t.Errorf("expected no clone calls on cache hit, got %d", len(fg.cloneURLs))
	}
}

func TestFetch_CacheMiss_PublicRepo(t *testing.T) {
	cacheRoot := t.TempDir()

	fg := newFakeGit()
	// Only HTTPS works (public repo, no SSH key)
	fg.allowedURLs["https://github.com/myorg/myrepo.git"] = true
	fg.allowedVersions["v1.0"] = true

	c := &Cache{
		Root:       cacheRoot,
		GitFunc:    fg.gitFunc,
		GetenvFunc: func(string) string { return "" },
	}

	path, err := c.Fetch("myorg", "myrepo", "v1.0")
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}

	expected := c.PackagePath("myorg", "myrepo", "v1.0")
	if path != expected {
		t.Errorf("Fetch() = %q, want %q", path, expected)
	}

	// Verify .git was removed
	if _, err := os.Stat(filepath.Join(path, ".git")); !os.IsNotExist(err) {
		t.Error("expected .git to be removed from cached package")
	}

	// Verify the cache entry exists
	if info, err := os.Stat(path); err != nil || !info.IsDir() {
		t.Errorf("expected cache directory to exist at %s", path)
	}
}

func TestFetch_CacheMiss_SSHSucceeds(t *testing.T) {
	cacheRoot := t.TempDir()

	fg := newFakeGit()
	fg.allowedURLs["git@github.com:myorg/myrepo.git"] = true
	fg.allowedVersions["v2.0"] = true

	c := &Cache{
		Root:       cacheRoot,
		GitFunc:    fg.gitFunc,
		GetenvFunc: func(string) string { return "" },
	}

	path, err := c.Fetch("myorg", "myrepo", "v2.0")
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}

	if !strings.Contains(path, "v2.0") {
		t.Errorf("expected path to contain v2.0, got %s", path)
	}

	// SSH should have been tried first and succeeded
	if len(fg.cloneURLs) != 1 {
		t.Errorf("expected 1 clone attempt (SSH), got %d: %v", len(fg.cloneURLs), fg.cloneURLs)
	}
	if fg.cloneURLs[0] != "git@github.com:myorg/myrepo.git" {
		t.Errorf("expected SSH URL first, got %s", fg.cloneURLs[0])
	}
}

func TestFetch_CacheMiss_TokenFallback(t *testing.T) {
	cacheRoot := t.TempDir()

	fg := newFakeGit()
	// SSH fails, token URL succeeds
	fg.allowedURLs["https://mytoken@github.com/myorg/myrepo.git"] = true
	fg.allowedVersions["v1.0"] = true

	c := &Cache{
		Root:    cacheRoot,
		GitFunc: fg.gitFunc,
		GetenvFunc: func(key string) string {
			if key == "GITHUB_TOKEN" {
				return "mytoken"
			}
			return ""
		},
	}

	path, err := c.Fetch("myorg", "myrepo", "v1.0")
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}
	if path == "" {
		t.Fatal("expected non-empty path")
	}

	// Should have tried SSH first, then token URL
	if len(fg.cloneURLs) != 2 {
		t.Fatalf("expected 2 clone attempts, got %d: %v", len(fg.cloneURLs), fg.cloneURLs)
	}
	if !strings.HasPrefix(fg.cloneURLs[0], "git@") {
		t.Errorf("first attempt should be SSH, got %s", fg.cloneURLs[0])
	}
	if !strings.Contains(fg.cloneURLs[1], "mytoken") {
		t.Errorf("second attempt should use token, got %s", fg.cloneURLs[1])
	}
}

func TestFetch_AllClonesFail(t *testing.T) {
	cacheRoot := t.TempDir()

	fg := newFakeGit()
	// No URLs succeed

	c := &Cache{
		Root:       cacheRoot,
		GitFunc:    fg.gitFunc,
		GetenvFunc: func(string) string { return "" },
	}

	_, err := c.Fetch("myorg", "myrepo", "v1.0")
	if err == nil {
		t.Fatal("expected error when all clones fail")
	}
	if !strings.Contains(err.Error(), "could not fetch myorg/myrepo") {
		t.Errorf("error message should mention repo: %v", err)
	}
	if !strings.Contains(err.Error(), "SSH key") {
		t.Errorf("error message should mention SSH: %v", err)
	}
	if !strings.Contains(err.Error(), "GITHUB_TOKEN") {
		t.Errorf("error message should mention GITHUB_TOKEN: %v", err)
	}
}

func TestFetch_InvalidVersion(t *testing.T) {
	cacheRoot := t.TempDir()

	fg := newFakeGit()
	fg.allowedURLs["https://github.com/myorg/myrepo.git"] = true
	// v999.0 is NOT in allowedVersions

	c := &Cache{
		Root:       cacheRoot,
		GitFunc:    fg.gitFunc,
		GetenvFunc: func(string) string { return "" },
	}

	_, err := c.Fetch("myorg", "myrepo", "v999.0")
	if err == nil {
		t.Fatal("expected error for invalid version")
	}
	if !strings.Contains(err.Error(), "v999.0") {
		t.Errorf("error should mention the version: %v", err)
	}
}

func TestFetch_MutableRef_Warning(t *testing.T) {
	cacheRoot := t.TempDir()

	fg := newFakeGit()
	fg.allowedURLs["https://github.com/myorg/myrepo.git"] = true
	fg.allowedVersions["main"] = true

	var buf bytes.Buffer
	c := &Cache{
		Root:       cacheRoot,
		GitFunc:    fg.gitFunc,
		WarnWriter: &buf,
		GetenvFunc: func(string) string { return "" },
	}

	path, err := c.Fetch("myorg", "myrepo", "main")
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}

	// Warning emitted
	if !strings.Contains(buf.String(), "mutable ref @main") {
		t.Errorf("expected mutable ref warning, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "reproducible") {
		t.Errorf("expected reproducibility warning, got: %s", buf.String())
	}

	// Path should include the date-stamped version
	if !strings.Contains(path, "main~") {
		t.Errorf("expected date-stamped cache path for mutable ref, got: %s", path)
	}
}

func TestFetch_MutableRef_NilWarnWriter(t *testing.T) {
	cacheRoot := t.TempDir()

	fg := newFakeGit()
	fg.allowedURLs["https://github.com/myorg/myrepo.git"] = true
	fg.allowedVersions["main"] = true

	c := &Cache{
		Root:       cacheRoot,
		GitFunc:    fg.gitFunc,
		WarnWriter: nil, // no warn writer
		GetenvFunc: func(string) string { return "" },
	}

	_, err := c.Fetch("myorg", "myrepo", "main")
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}
	// No panic — nil WarnWriter is handled
}

func TestFetch_AtomicWrite_NoPartialEntry(t *testing.T) {
	cacheRoot := t.TempDir()

	fg := newFakeGit()
	fg.allowedURLs["https://github.com/myorg/myrepo.git"] = true
	// Checkout fails — should not leave a partial cache entry

	c := &Cache{
		Root:       cacheRoot,
		GitFunc:    fg.gitFunc,
		GetenvFunc: func(string) string { return "" },
	}

	_, err := c.Fetch("myorg", "myrepo", "v1.0")
	if err == nil {
		t.Fatal("expected error")
	}

	// No cache entry should exist
	target := c.PackagePath("myorg", "myrepo", "v1.0")
	if _, statErr := os.Stat(target); !os.IsNotExist(statErr) {
		t.Errorf("expected no cache entry at %s after failed fetch", target)
	}

	// No temp dirs should be left
	entries, _ := os.ReadDir(filepath.Dir(target))
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".fetch-") {
			t.Errorf("found leftover temp dir: %s", e.Name())
		}
	}
}

func TestFetch_SecondCall_CacheHit(t *testing.T) {
	cacheRoot := t.TempDir()

	fg := newFakeGit()
	fg.allowedURLs["https://github.com/myorg/myrepo.git"] = true
	fg.allowedVersions["v1.0"] = true

	c := &Cache{
		Root:       cacheRoot,
		GitFunc:    fg.gitFunc,
		GetenvFunc: func(string) string { return "" },
	}

	// First fetch — cache miss
	path1, err := c.Fetch("myorg", "myrepo", "v1.0")
	if err != nil {
		t.Fatalf("first Fetch() error: %v", err)
	}

	cloneCount := len(fg.cloneURLs)

	// Second fetch — cache hit
	path2, err := c.Fetch("myorg", "myrepo", "v1.0")
	if err != nil {
		t.Fatalf("second Fetch() error: %v", err)
	}

	if path1 != path2 {
		t.Errorf("paths differ between calls: %q vs %q", path1, path2)
	}
	if len(fg.cloneURLs) != cloneCount {
		t.Errorf("expected no new clone calls on cache hit, got %d new", len(fg.cloneURLs)-cloneCount)
	}
}

func TestFetch_WithRepoFiles(t *testing.T) {
	cacheRoot := t.TempDir()

	fg := newFakeGit()
	fg.allowedURLs["https://github.com/myorg/myrepo.git"] = true
	fg.allowedVersions["v1.0"] = true
	fg.repoFiles[".pi/docker/up.yaml"] = "bash: docker compose up -d"
	fg.repoFiles["pi-package.yaml"] = "min_pi_version: \"0.1\""

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

	// Check that repo files are present
	content, err := os.ReadFile(filepath.Join(path, ".pi", "docker", "up.yaml"))
	if err != nil {
		t.Fatalf("reading cached file: %v", err)
	}
	if string(content) != "bash: docker compose up -d" {
		t.Errorf("unexpected file content: %s", string(content))
	}
}

func TestFetch_PrivateRepo_AuthInstructions(t *testing.T) {
	cacheRoot := t.TempDir()

	fg := newFakeGit()
	// No URLs succeed — simulates a private repo with no auth

	c := &Cache{
		Root:       cacheRoot,
		GitFunc:    fg.gitFunc,
		GetenvFunc: func(string) string { return "" },
	}

	_, err := c.Fetch("private-org", "secret-repo", "v1.0")
	if err == nil {
		t.Fatal("expected error for private repo with no auth")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "private-org/secret-repo") {
		t.Errorf("error should mention org/repo: %v", err)
	}
	if !strings.Contains(errMsg, "SSH key") {
		t.Errorf("error should mention SSH: %v", err)
	}
	if !strings.Contains(errMsg, "GITHUB_TOKEN") {
		t.Errorf("error should mention GITHUB_TOKEN: %v", err)
	}
}

func TestFetch_CloneOrder_SSHThenTokenThenHTTPS(t *testing.T) {
	cacheRoot := t.TempDir()

	fg := newFakeGit()
	// Only plain HTTPS works
	fg.allowedURLs["https://github.com/org/repo.git"] = true
	fg.allowedVersions["v1.0"] = true

	c := &Cache{
		Root:    cacheRoot,
		GitFunc: fg.gitFunc,
		GetenvFunc: func(key string) string {
			if key == "GITHUB_TOKEN" {
				return "tok123"
			}
			return ""
		},
	}

	_, err := c.Fetch("org", "repo", "v1.0")
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}

	// Should have tried: SSH → token HTTPS → plain HTTPS
	if len(fg.cloneURLs) != 3 {
		t.Fatalf("expected 3 clone attempts, got %d: %v", len(fg.cloneURLs), fg.cloneURLs)
	}
	if !strings.HasPrefix(fg.cloneURLs[0], "git@") {
		t.Errorf("first should be SSH: %s", fg.cloneURLs[0])
	}
	if !strings.Contains(fg.cloneURLs[1], "tok123") {
		t.Errorf("second should use token: %s", fg.cloneURLs[1])
	}
	if fg.cloneURLs[2] != "https://github.com/org/repo.git" {
		t.Errorf("third should be plain HTTPS: %s", fg.cloneURLs[2])
	}
}

func TestFetch_NoToken_SkipsTokenAttempt(t *testing.T) {
	cacheRoot := t.TempDir()

	fg := newFakeGit()
	fg.allowedURLs["https://github.com/org/repo.git"] = true
	fg.allowedVersions["v1.0"] = true

	c := &Cache{
		Root:       cacheRoot,
		GitFunc:    fg.gitFunc,
		GetenvFunc: func(string) string { return "" },
	}

	_, err := c.Fetch("org", "repo", "v1.0")
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}

	// SSH → plain HTTPS (no token attempt)
	if len(fg.cloneURLs) != 2 {
		t.Fatalf("expected 2 clone attempts (no token), got %d: %v", len(fg.cloneURLs), fg.cloneURLs)
	}
}
