package refparser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RefType identifies the kind of automation reference.
type RefType int

const (
	RefLocal   RefType = iota // .pi/ local automation
	RefBuiltin                // pi:name built-in
	RefGitHub                 // org/repo@version/path
	RefFile                   // file:~/path or file:/abs/path
	RefAlias                  // alias/path (resolved via packages:)
)

// AutomationRef is the parsed representation of an automation reference string.
// All fields relevant to the reference type are populated; others are empty.
type AutomationRef struct {
	Type    RefType
	Raw     string // original input string
	Path    string // automation path within the source (e.g. "docker/up")
	Org     string // GitHub org (GitHubRef only)
	Repo    string // GitHub repo (GitHubRef only)
	Version string // version tag (GitHubRef only, e.g. "v1.2")
	FSPath  string // filesystem path (FileRef only, with ~ expanded)
	Alias   string // alias name (AliasRef only)
}

// String returns the canonical string form of the reference.
func (r AutomationRef) String() string {
	switch r.Type {
	case RefBuiltin:
		return "pi:" + r.Path
	case RefGitHub:
		s := r.Org + "/" + r.Repo + "@" + r.Version
		if r.Path != "" {
			s += "/" + r.Path
		}
		return s
	case RefFile:
		return "file:" + r.FSPath
		// Path is derived from FSPath; not included separately in the canonical form
	case RefAlias:
		if r.Path != "" {
			return r.Alias + "/" + r.Path
		}
		return r.Alias
	default:
		return r.Path
	}
}

// Parse classifies a raw automation reference string into a typed AutomationRef.
// It performs no filesystem access or network calls — pure string parsing.
//
// Known aliases is the set of alias names declared in pi.yaml packages: block.
// Pass nil if no aliases are configured.
func Parse(ref string, knownAliases map[string]bool) (AutomationRef, error) {
	if ref == "" {
		return AutomationRef{}, fmt.Errorf("empty automation reference")
	}

	ref = strings.TrimSpace(ref)

	// 1. Built-in: starts with "pi:"
	if strings.HasPrefix(ref, "pi:") {
		path := strings.TrimPrefix(ref, "pi:")
		if path == "" {
			return AutomationRef{}, fmt.Errorf("invalid automation reference %q: missing name after \"pi:\"", ref)
		}
		return AutomationRef{
			Type: RefBuiltin,
			Raw:  ref,
			Path: normalizePath(path),
		}, nil
	}

	// 2. File source: starts with "file:"
	if strings.HasPrefix(ref, "file:") {
		fsPath := strings.TrimPrefix(ref, "file:")
		if fsPath == "" {
			return AutomationRef{}, fmt.Errorf("invalid automation reference %q: missing path after \"file:\"", ref)
		}
		expanded := expandTilde(fsPath)
		return AutomationRef{
			Type:   RefFile,
			Raw:    ref,
			FSPath: expanded,
			Path:   extractFileRefPath(expanded),
		}, nil
	}

	// 3. GitHub ref: contains "@"
	if idx := strings.Index(ref, "@"); idx >= 0 {
		if idx == 0 {
			return AutomationRef{}, fmt.Errorf("invalid automation reference %q: missing org/repo before \"@\"", ref)
		}
		return parseGitHubRef(ref, idx)
	}

	// 4. Alias ref: first path segment matches a known alias
	if len(knownAliases) > 0 {
		firstSeg := firstSegment(ref)
		if knownAliases[firstSeg] {
			rest := strings.TrimPrefix(ref, firstSeg)
			rest = strings.TrimPrefix(rest, "/")
			return AutomationRef{
				Type:  RefAlias,
				Raw:   ref,
				Alias: firstSeg,
				Path:  normalizePath(rest),
			}, nil
		}
	}

	// 5. Default: local .pi/ reference
	return AutomationRef{
		Type: RefLocal,
		Raw:  ref,
		Path: normalizePath(ref),
	}, nil
}

// parseGitHubRef parses "org/repo@version" or "org/repo@version/path".
func parseGitHubRef(ref string, atIdx int) (AutomationRef, error) {
	prefix := ref[:atIdx]   // "org/repo"
	suffix := ref[atIdx+1:] // "version" or "version/path"

	parts := strings.SplitN(prefix, "/", 2)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return AutomationRef{}, fmt.Errorf("invalid automation reference %q: GitHub references must be \"org/repo@version[/path]\"", ref)
	}

	org := parts[0]
	repo := parts[1]

	if strings.Contains(repo, "/") {
		return AutomationRef{}, fmt.Errorf("invalid automation reference %q: GitHub references must be \"org/repo@version[/path]\" — repo name cannot contain \"/\"", ref)
	}

	if suffix == "" {
		return AutomationRef{}, fmt.Errorf("invalid automation reference %q: missing version after \"@\"", ref)
	}

	var version, path string
	if slashIdx := strings.Index(suffix, "/"); slashIdx >= 0 {
		version = suffix[:slashIdx]
		path = suffix[slashIdx+1:]
	} else {
		version = suffix
	}

	if version == "" {
		return AutomationRef{}, fmt.Errorf("invalid automation reference %q: missing version after \"@\"", ref)
	}

	return AutomationRef{
		Type:    RefGitHub,
		Raw:     ref,
		Org:     org,
		Repo:    repo,
		Version: version,
		Path:    normalizePath(path),
	}, nil
}

// firstSegment returns the part of a path before the first "/".
func firstSegment(s string) string {
	if idx := strings.Index(s, "/"); idx >= 0 {
		return s[:idx]
	}
	return s
}

// normalizePath lowercases and trims slashes.
func normalizePath(p string) string {
	p = strings.ToLower(p)
	p = strings.Trim(p, "/")
	return filepath.ToSlash(p)
}

// expandTilde replaces a leading "~/" with the user's home directory.
func expandTilde(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

// extractFileRefPath derives the automation path from a file: source filesystem path.
// It uses the last path component (without extension) as a simple default.
func extractFileRefPath(fsPath string) string {
	base := filepath.Base(fsPath)
	ext := filepath.Ext(base)
	if ext != "" {
		base = strings.TrimSuffix(base, ext)
	}
	return normalizePath(base)
}
