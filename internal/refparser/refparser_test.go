package refparser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseLocalRef(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantPath string
	}{
		{"simple", "docker/up", "docker/up"},
		{"nested", "setup/cursor/install", "setup/cursor/install"},
		{"single", "build", "build"},
		{"trailing slash", "docker/up/", "docker/up"},
		{"leading slash", "/docker/up", "docker/up"},
		{"mixed case", "Docker/Up", "docker/up"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref, err := Parse(tt.input, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ref.Type != RefLocal {
				t.Errorf("type = %v, want RefLocal", ref.Type)
			}
			if ref.Path != tt.wantPath {
				t.Errorf("path = %q, want %q", ref.Path, tt.wantPath)
			}
			if ref.Raw != tt.input {
				t.Errorf("raw = %q, want %q", ref.Raw, tt.input)
			}
		})
	}
}

func TestParseBuiltinRef(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantPath string
	}{
		{"simple", "pi:install-go", "install-go"},
		{"nested", "pi:docker/up", "docker/up"},
		{"deep", "pi:cursor/install-extensions", "cursor/install-extensions"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref, err := Parse(tt.input, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ref.Type != RefBuiltin {
				t.Errorf("type = %v, want RefBuiltin", ref.Type)
			}
			if ref.Path != tt.wantPath {
				t.Errorf("path = %q, want %q", ref.Path, tt.wantPath)
			}
		})
	}
}

func TestParseBuiltinRefErrors(t *testing.T) {
	_, err := Parse("pi:", nil)
	if err == nil {
		t.Fatal("expected error for empty builtin name")
	}
}

func TestParseGitHubRef(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantOrg     string
		wantRepo    string
		wantVersion string
		wantPath    string
	}{
		{
			"full",
			"yotam180/pi-common@v1.2/docker/up",
			"yotam180", "pi-common", "v1.2", "docker/up",
		},
		{
			"no path",
			"yotam180/pi-common@v1.2",
			"yotam180", "pi-common", "v1.2", "",
		},
		{
			"main branch",
			"org/repo@main/setup/install",
			"org", "repo", "main", "setup/install",
		},
		{
			"deep path",
			"company/shared@v2.0.0/setup/install-go",
			"company", "shared", "v2.0.0", "setup/install-go",
		},
		{
			"semver with patch",
			"acme/tools@v1.2.3/build",
			"acme", "tools", "v1.2.3", "build",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref, err := Parse(tt.input, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ref.Type != RefGitHub {
				t.Errorf("type = %v, want RefGitHub", ref.Type)
			}
			if ref.Org != tt.wantOrg {
				t.Errorf("org = %q, want %q", ref.Org, tt.wantOrg)
			}
			if ref.Repo != tt.wantRepo {
				t.Errorf("repo = %q, want %q", ref.Repo, tt.wantRepo)
			}
			if ref.Version != tt.wantVersion {
				t.Errorf("version = %q, want %q", ref.Version, tt.wantVersion)
			}
			if ref.Path != tt.wantPath {
				t.Errorf("path = %q, want %q", ref.Path, tt.wantPath)
			}
		})
	}
}

func TestParseGitHubRefErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"missing org", "/@v1.0/path"},
		{"missing repo", "org/@v1.0/path"},
		{"missing version", "org/repo@"},
		{"missing version with path", "org/repo@/path"},
		{"no org prefix", "@v1.0/path"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.input, nil)
			if err == nil {
				t.Fatalf("expected error for %q", tt.input)
			}
		})
	}
}

func TestParseFileRef(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot get home directory")
	}

	tests := []struct {
		name       string
		input      string
		wantFSPath string
	}{
		{
			"tilde",
			"file:~/my-automations/docker/up",
			filepath.Join(home, "my-automations/docker/up"),
		},
		{
			"absolute",
			"file:/opt/automations/build",
			"/opt/automations/build",
		},
		{
			"relative",
			"file:../shared/deploy",
			"../shared/deploy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref, err := Parse(tt.input, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ref.Type != RefFile {
				t.Errorf("type = %v, want RefFile", ref.Type)
			}
			if ref.FSPath != tt.wantFSPath {
				t.Errorf("fspath = %q, want %q", ref.FSPath, tt.wantFSPath)
			}
		})
	}
}

func TestParseFileRefError(t *testing.T) {
	_, err := Parse("file:", nil)
	if err == nil {
		t.Fatal("expected error for empty file path")
	}
}

func TestParseAliasRef(t *testing.T) {
	aliases := map[string]bool{
		"common": true,
		"shared": true,
	}

	tests := []struct {
		name      string
		input     string
		wantAlias string
		wantPath  string
	}{
		{
			"with path",
			"common/docker/up",
			"common", "docker/up",
		},
		{
			"single segment",
			"common",
			"common", "",
		},
		{
			"different alias",
			"shared/setup/install-go",
			"shared", "setup/install-go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref, err := Parse(tt.input, aliases)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ref.Type != RefAlias {
				t.Errorf("type = %v, want RefAlias", ref.Type)
			}
			if ref.Alias != tt.wantAlias {
				t.Errorf("alias = %q, want %q", ref.Alias, tt.wantAlias)
			}
			if ref.Path != tt.wantPath {
				t.Errorf("path = %q, want %q", ref.Path, tt.wantPath)
			}
		})
	}
}

func TestAliasNotMatchedFallsToLocal(t *testing.T) {
	aliases := map[string]bool{
		"common": true,
	}

	ref, err := Parse("docker/up", aliases)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref.Type != RefLocal {
		t.Errorf("type = %v, want RefLocal (non-alias should fall through)", ref.Type)
	}
}

func TestEmptyRefError(t *testing.T) {
	_, err := Parse("", nil)
	if err == nil {
		t.Fatal("expected error for empty reference")
	}
}

func TestWhitespaceHandling(t *testing.T) {
	ref, err := Parse("  docker/up  ", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref.Path != "docker/up" {
		t.Errorf("path = %q, want %q", ref.Path, "docker/up")
	}
}

func TestStringRoundTrip(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		aliases   map[string]bool
		wantStr   string
	}{
		{"local", "docker/up", nil, "docker/up"},
		{"builtin", "pi:install-go", nil, "pi:install-go"},
		{"github with path", "org/repo@v1.0/docker/up", nil, "org/repo@v1.0/docker/up"},
		{"github no path", "org/repo@v1.0", nil, "org/repo@v1.0"},
		{"alias with path", "common/docker/up", map[string]bool{"common": true}, "common/docker/up"},
		{"alias no path", "common", map[string]bool{"common": true}, "common"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref, err := Parse(tt.input, tt.aliases)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := ref.String()
			if got != tt.wantStr {
				t.Errorf("String() = %q, want %q", got, tt.wantStr)
			}
		})
	}
}

func TestFileRefString(t *testing.T) {
	ref, err := Parse("file:~/my-automations/docker/up", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := ref.String()
	if s == "" {
		t.Error("String() returned empty for file ref")
	}
	if ref.Type != RefFile {
		t.Errorf("type = %v, want RefFile", ref.Type)
	}
}

func TestParsePrecedence(t *testing.T) {
	// "pi:" wins over everything
	ref, err := Parse("pi:something@v1.0", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref.Type != RefBuiltin {
		t.Errorf("pi: prefix should win; got type %v", ref.Type)
	}

	// "file:" wins over @ detection
	ref, err = Parse("file:some@path", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref.Type != RefFile {
		t.Errorf("file: prefix should win; got type %v", ref.Type)
	}

	// Alias wins over local when first segment matches
	aliases := map[string]bool{"docker": true}
	ref, err = Parse("docker/up", aliases)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref.Type != RefAlias {
		t.Errorf("alias should win over local when matched; got type %v", ref.Type)
	}
}

func TestGitHubRefRepoSlashRejected(t *testing.T) {
	// "org/repo/extra@v1" — repo contains "/"
	// The prefix before @ is "org/repo/extra", SplitN("/", 2) gives ["org", "repo/extra"]
	// repo = "repo/extra" which contains "/", should be rejected
	_, err := Parse("org/repo/extra@v1", nil)
	if err == nil {
		t.Fatal("expected error for repo name containing /")
	}
}
