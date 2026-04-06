package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vyper-tooling/pi/internal/automation"
	"github.com/vyper-tooling/pi/internal/config"
	"github.com/vyper-tooling/pi/internal/display"
	"github.com/vyper-tooling/pi/internal/discovery"
	"github.com/vyper-tooling/pi/internal/refparser"
)

func emptyResult() *discovery.Result {
	return discovery.NewResult(make(map[string]*automation.Automation), nil)
}

// mockFetcher implements PackageFetcher for testing.
type mockFetcher struct {
	result    map[string]string // "org/repo@version" → path
	wasCached map[string]bool
	err       error
	calls     []string
}

func newMockFetcher() *mockFetcher {
	return &mockFetcher{
		result:    make(map[string]string),
		wasCached: make(map[string]bool),
	}
}

func (m *mockFetcher) add(org, repo, version, path string, cached bool) {
	key := org + "/" + repo + "@" + version
	m.result[key] = path
	m.wasCached[key] = cached
}

func (m *mockFetcher) Fetch(org, repo, version string) (string, bool, error) {
	key := org + "/" + repo + "@" + version
	m.calls = append(m.calls, key)
	if m.err != nil {
		return "", false, m.err
	}
	path, ok := m.result[key]
	if !ok {
		return "", false, fmt.Errorf("package %s not found", key)
	}
	return path, m.wasCached[key], nil
}

// --- resolveGitHubPackage tests ---

func TestResolveGitHubPackage_CacheHit(t *testing.T) {
	pkgDir := t.TempDir()
	fetcher := newMockFetcher()
	fetcher.add("acme", "tools", "v1.0", pkgDir, true)

	var buf bytes.Buffer
	printer := display.NewForWriter(&buf)
	pkg := config.PackageEntry{Source: "acme/tools@v1.0"}

	got, err := resolveGitHubPackage(pkg, &buf, printer, fetcher)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != pkgDir {
		t.Errorf("got %q, want %q", got, pkgDir)
	}
	if !strings.Contains(buf.String(), "cached") {
		t.Errorf("expected 'cached' status in output, got: %q", buf.String())
	}
}

func TestResolveGitHubPackage_FreshFetch(t *testing.T) {
	pkgDir := t.TempDir()
	fetcher := newMockFetcher()
	fetcher.add("acme", "tools", "v2.0", pkgDir, false)

	var buf bytes.Buffer
	printer := display.NewForWriter(&buf)
	pkg := config.PackageEntry{Source: "acme/tools@v2.0"}

	got, err := resolveGitHubPackage(pkg, &buf, printer, fetcher)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != pkgDir {
		t.Errorf("got %q, want %q", got, pkgDir)
	}
	if !strings.Contains(buf.String(), "fetched") {
		t.Errorf("expected 'fetched' status in output, got: %q", buf.String())
	}
}

func TestResolveGitHubPackage_FetchError(t *testing.T) {
	fetcher := newMockFetcher()
	fetcher.err = fmt.Errorf("network timeout")

	var buf bytes.Buffer
	printer := display.NewForWriter(&buf)
	pkg := config.PackageEntry{Source: "acme/tools@v1.0"}

	_, err := resolveGitHubPackage(pkg, &buf, printer, fetcher)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "network timeout") {
		t.Errorf("error should contain cause, got: %v", err)
	}
	if !strings.Contains(buf.String(), "failed") {
		t.Errorf("expected 'failed' status in output, got: %q", buf.String())
	}
}

func TestResolveGitHubPackage_InvalidSource(t *testing.T) {
	fetcher := newMockFetcher()

	var buf bytes.Buffer
	printer := display.NewForWriter(&buf)
	pkg := config.PackageEntry{Source: "not-a-valid-ref"}

	_, err := resolveGitHubPackage(pkg, &buf, printer, fetcher)
	if err == nil {
		t.Fatal("expected error for invalid source")
	}
	if !strings.Contains(err.Error(), "invalid package source") {
		t.Errorf("error should mention invalid source, got: %v", err)
	}
}

func TestResolveGitHubPackage_NilPrinter(t *testing.T) {
	pkgDir := t.TempDir()
	fetcher := newMockFetcher()
	fetcher.add("acme", "tools", "v1.0", pkgDir, true)

	pkg := config.PackageEntry{Source: "acme/tools@v1.0"}

	got, err := resolveGitHubPackage(pkg, nil, nil, fetcher)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != pkgDir {
		t.Errorf("got %q, want %q", got, pkgDir)
	}
}

// --- fetchGitHubPackage tests ---

func TestFetchGitHubPackage_Cached(t *testing.T) {
	pkgDir := t.TempDir()
	fetcher := newMockFetcher()
	fetcher.add("acme", "tools", "v1.0", pkgDir, true)

	var buf bytes.Buffer
	printer := display.NewForWriter(&buf)
	ref := refparser.AutomationRef{Org: "acme", Repo: "tools", Version: "v1.0", Type: refparser.RefGitHub}

	err := fetchGitHubPackage(ref, &buf, printer, fetcher)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "cached") {
		t.Errorf("expected 'cached' in output, got: %q", buf.String())
	}
}

func TestFetchGitHubPackage_Fresh(t *testing.T) {
	pkgDir := t.TempDir()
	fetcher := newMockFetcher()
	fetcher.add("acme", "tools", "v2.0", pkgDir, false)

	var buf bytes.Buffer
	printer := display.NewForWriter(&buf)
	ref := refparser.AutomationRef{Org: "acme", Repo: "tools", Version: "v2.0", Type: refparser.RefGitHub}

	err := fetchGitHubPackage(ref, &buf, printer, fetcher)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "fetched") {
		t.Errorf("expected 'fetched' in output, got: %q", buf.String())
	}
}

func TestFetchGitHubPackage_NoVersion(t *testing.T) {
	fetcher := newMockFetcher()

	var buf bytes.Buffer
	printer := display.NewForWriter(&buf)
	ref := refparser.AutomationRef{Org: "acme", Repo: "tools", Version: "", Type: refparser.RefGitHub}

	err := fetchGitHubPackage(ref, &buf, printer, fetcher)
	if err == nil {
		t.Fatal("expected error for missing version")
	}
	if !strings.Contains(err.Error(), "version required") {
		t.Errorf("error should mention 'version required', got: %v", err)
	}
}

func TestFetchGitHubPackage_FetchError(t *testing.T) {
	fetcher := newMockFetcher()
	fetcher.err = fmt.Errorf("auth failed")

	var buf bytes.Buffer
	printer := display.NewForWriter(&buf)
	ref := refparser.AutomationRef{Org: "acme", Repo: "tools", Version: "v1.0", Type: refparser.RefGitHub}

	err := fetchGitHubPackage(ref, &buf, printer, fetcher)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "auth failed") {
		t.Errorf("error should contain cause, got: %v", err)
	}
	if !strings.Contains(buf.String(), "failed") {
		t.Errorf("expected 'failed' in output, got: %q", buf.String())
	}
}

// --- newOnDemandFetcher tests ---

func TestOnDemandFetcher_FetchAndMerge(t *testing.T) {
	pkgDir := t.TempDir()
	piDir := filepath.Join(pkgDir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "deploy.yaml"), []byte("description: deploy\nbash: echo deploy\n"), 0644)

	fetcher := newMockFetcher()
	fetcher.add("acme", "ops", "v1.0", pkgDir, false)

	var buf bytes.Buffer
	fetchFn := newOnDemandFetcher(&buf, fetcher)

	result := emptyResult()
	ref := refparser.AutomationRef{
		Org: "acme", Repo: "ops", Version: "v1.0", Path: "deploy",
		Type: refparser.RefGitHub, Raw: "acme/ops@v1.0/deploy",
	}

	a, err := fetchFn(result, ref)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a == nil {
		t.Fatal("expected automation, got nil")
	}
	if a.Description != "deploy" {
		t.Errorf("description = %q, want %q", a.Description, "deploy")
	}

	if !strings.Contains(buf.String(), "fetched (on demand)") {
		t.Errorf("expected advisory for fresh fetch, got: %q", buf.String())
	}
	if !strings.Contains(buf.String(), "packages:") {
		t.Errorf("expected packages snippet in advisory, got: %q", buf.String())
	}
}

func TestOnDemandFetcher_CachedHitNoAdvisory(t *testing.T) {
	pkgDir := t.TempDir()
	piDir := filepath.Join(pkgDir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "build.yaml"), []byte("description: build\nbash: echo build\n"), 0644)

	fetcher := newMockFetcher()
	fetcher.add("acme", "ops", "v1.0", pkgDir, true)

	var buf bytes.Buffer
	fetchFn := newOnDemandFetcher(&buf, fetcher)

	result := emptyResult()
	ref := refparser.AutomationRef{
		Org: "acme", Repo: "ops", Version: "v1.0", Path: "build",
		Type: refparser.RefGitHub, Raw: "acme/ops@v1.0/build",
	}

	a, err := fetchFn(result, ref)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a == nil {
		t.Fatal("expected automation, got nil")
	}

	if strings.Contains(buf.String(), "tip:") {
		t.Errorf("cached hit should not print advisory, got: %q", buf.String())
	}
}

func TestOnDemandFetcher_Dedup(t *testing.T) {
	pkgDir := t.TempDir()
	piDir := filepath.Join(pkgDir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "a.yaml"), []byte("description: a\nbash: echo a\n"), 0644)
	os.WriteFile(filepath.Join(piDir, "b.yaml"), []byte("description: b\nbash: echo b\n"), 0644)

	fetcher := newMockFetcher()
	fetcher.add("acme", "ops", "v1.0", pkgDir, false)

	var buf bytes.Buffer
	fetchFn := newOnDemandFetcher(&buf, fetcher)

	result := emptyResult()

	refA := refparser.AutomationRef{
		Org: "acme", Repo: "ops", Version: "v1.0", Path: "a",
		Type: refparser.RefGitHub, Raw: "acme/ops@v1.0/a",
	}
	refB := refparser.AutomationRef{
		Org: "acme", Repo: "ops", Version: "v1.0", Path: "b",
		Type: refparser.RefGitHub, Raw: "acme/ops@v1.0/b",
	}

	if _, err := fetchFn(result, refA); err != nil {
		t.Fatalf("first call error: %v", err)
	}
	if _, err := fetchFn(result, refB); err != nil {
		t.Fatalf("second call error: %v", err)
	}

	if len(fetcher.calls) != 1 {
		t.Errorf("fetcher should be called once (deduped), got %d calls", len(fetcher.calls))
	}
}

func TestOnDemandFetcher_FetchError(t *testing.T) {
	fetcher := newMockFetcher()
	fetcher.err = fmt.Errorf("connection refused")

	var buf bytes.Buffer
	fetchFn := newOnDemandFetcher(&buf, fetcher)

	result := emptyResult()
	ref := refparser.AutomationRef{
		Org: "acme", Repo: "ops", Version: "v1.0", Path: "deploy",
		Type: refparser.RefGitHub, Raw: "acme/ops@v1.0/deploy",
	}

	_, err := fetchFn(result, ref)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("error should contain cause, got: %v", err)
	}
	if !strings.Contains(err.Error(), "on-demand fetch") {
		t.Errorf("error should mention on-demand fetch, got: %v", err)
	}
}

func TestOnDemandFetcher_EmptyPath(t *testing.T) {
	pkgDir := t.TempDir()
	piDir := filepath.Join(pkgDir, ".pi")
	os.MkdirAll(piDir, 0755)

	fetcher := newMockFetcher()
	fetcher.add("acme", "ops", "v1.0", pkgDir, true)

	fetchFn := newOnDemandFetcher(nil, fetcher)

	result := emptyResult()
	ref := refparser.AutomationRef{
		Org: "acme", Repo: "ops", Version: "v1.0", Path: "",
		Type: refparser.RefGitHub, Raw: "acme/ops@v1.0",
	}

	_, err := fetchFn(result, ref)
	if err == nil {
		t.Fatal("expected error for empty path")
	}
	if !strings.Contains(err.Error(), "no automation path") {
		t.Errorf("error should mention missing path, got: %v", err)
	}
}

func TestOnDemandFetcher_AutomationNotFound(t *testing.T) {
	pkgDir := t.TempDir()
	piDir := filepath.Join(pkgDir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "exists.yaml"), []byte("description: exists\nbash: echo\n"), 0644)

	fetcher := newMockFetcher()
	fetcher.add("acme", "ops", "v1.0", pkgDir, true)

	fetchFn := newOnDemandFetcher(nil, fetcher)

	result := emptyResult()
	ref := refparser.AutomationRef{
		Org: "acme", Repo: "ops", Version: "v1.0", Path: "nonexistent",
		Type: refparser.RefGitHub, Raw: "acme/ops@v1.0/nonexistent",
	}

	_, err := fetchFn(result, ref)
	if err == nil {
		t.Fatal("expected error for nonexistent automation")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention not found, got: %v", err)
	}
}

func TestOnDemandFetcher_NilStderr(t *testing.T) {
	pkgDir := t.TempDir()
	piDir := filepath.Join(pkgDir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "hello.yaml"), []byte("description: hello\nbash: echo hi\n"), 0644)

	fetcher := newMockFetcher()
	fetcher.add("acme", "ops", "v1.0", pkgDir, false)

	fetchFn := newOnDemandFetcher(nil, fetcher)

	result := emptyResult()
	ref := refparser.AutomationRef{
		Org: "acme", Repo: "ops", Version: "v1.0", Path: "hello",
		Type: refparser.RefGitHub, Raw: "acme/ops@v1.0/hello",
	}

	a, err := fetchFn(result, ref)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a == nil {
		t.Fatal("expected automation, got nil")
	}
}

// --- mergePackages tests with fetcher ---

func TestMergePackages_GitHubPackage(t *testing.T) {
	root := t.TempDir()
	piDir := filepath.Join(root, ".pi")
	os.MkdirAll(piDir, 0755)

	pkgDir := t.TempDir()
	pkgPiDir := filepath.Join(pkgDir, ".pi")
	os.MkdirAll(pkgPiDir, 0755)
	os.WriteFile(filepath.Join(pkgPiDir, "tool.yaml"), []byte("description: a tool\nbash: echo tool\n"), 0644)

	fetcher := newMockFetcher()
	fetcher.add("acme", "tools", "v1.0", pkgDir, true)

	cfg := &config.ProjectConfig{
		Project: "test",
		Packages: []config.PackageEntry{
			{Source: "acme/tools@v1.0"},
		},
	}

	result := emptyResult()
	var buf bytes.Buffer
	err := mergePackages(result, cfg, root, &buf, fetcher)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	autos := result.PackageAutomations("acme/tools@v1.0")
	if len(autos) == 0 {
		t.Error("expected package automations to be merged")
	}
}

func TestMergePackages_GitHubFetchError(t *testing.T) {
	root := t.TempDir()
	piDir := filepath.Join(root, ".pi")
	os.MkdirAll(piDir, 0755)

	fetcher := newMockFetcher()
	fetcher.err = fmt.Errorf("git clone failed")

	cfg := &config.ProjectConfig{
		Project: "test",
		Packages: []config.PackageEntry{
			{Source: "acme/tools@v1.0"},
		},
	}

	result := emptyResult()
	var buf bytes.Buffer
	err := mergePackages(result, cfg, root, &buf, fetcher)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "git clone failed") {
		t.Errorf("error should contain cause, got: %v", err)
	}
}

// --- runAddWithFetcher tests ---

func TestRunAddWithFetcher_GitHubCached(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "pi.yaml", "project: test\n")

	pkgDir := t.TempDir()
	fetcher := newMockFetcher()
	fetcher.add("acme", "tools", "v1.0", pkgDir, true)

	var stdout, stderr bytes.Buffer
	err := runAddWithFetcher(dir, "acme/tools@v1.0", "", &stdout, &stderr, fetcher)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("reload error: %v", err)
	}
	if len(cfg.Packages) != 1 {
		t.Fatalf("packages count = %d, want 1", len(cfg.Packages))
	}
	if cfg.Packages[0].Source != "acme/tools@v1.0" {
		t.Errorf("source = %q, want %q", cfg.Packages[0].Source, "acme/tools@v1.0")
	}
	if !strings.Contains(stderr.String(), "added") {
		t.Errorf("stderr should contain 'added', got: %q", stderr.String())
	}
}

func TestRunAddWithFetcher_GitHubFetchError(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "pi.yaml", "project: test\n")

	fetcher := newMockFetcher()
	fetcher.err = fmt.Errorf("unauthorized")

	var stdout, stderr bytes.Buffer
	err := runAddWithFetcher(dir, "acme/tools@v1.0", "", &stdout, &stderr, fetcher)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "unauthorized") {
		t.Errorf("error should contain cause, got: %v", err)
	}
}

func TestRunAddWithFetcher_GitHubWithAlias(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "pi.yaml", "project: test\n")

	pkgDir := t.TempDir()
	fetcher := newMockFetcher()
	fetcher.add("acme", "tools", "v1.0", pkgDir, false)

	var stdout, stderr bytes.Buffer
	err := runAddWithFetcher(dir, "acme/tools@v1.0", "mytools", &stdout, &stderr, fetcher)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("reload error: %v", err)
	}
	if cfg.Packages[0].As != "mytools" {
		t.Errorf("alias = %q, want %q", cfg.Packages[0].As, "mytools")
	}
}

func TestRunAddWithFetcher_FileSourceIgnoresFetcher(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "pi.yaml", "project: test\n")

	fetcher := newMockFetcher()
	fetcher.err = fmt.Errorf("should not be called")

	var stdout, stderr bytes.Buffer
	err := runAddWithFetcher(dir, "file:~/path", "", &stdout, &stderr, fetcher)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fetcher.calls) != 0 {
		t.Errorf("fetcher should not be called for file: sources, got %d calls", len(fetcher.calls))
	}
}
