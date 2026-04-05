package cache

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Cache manages the PI package cache at ~/.pi/cache/.
type Cache struct {
	// Root is the cache root directory (defaults to ~/.pi/cache).
	Root string

	// WarnWriter receives warning messages (e.g. mutable ref warnings).
	// If nil, warnings are discarded.
	WarnWriter io.Writer

	// PIVersion is the running PI binary version for min_pi_version checks.
	// Empty string skips version checks.
	PIVersion string

	// GitFunc executes a git command and returns stdout, stderr, and error.
	// If nil, defaults to running the real git binary.
	GitFunc func(args []string, dir string) (stdout string, stderr string, err error)

	// GetenvFunc reads an environment variable. If nil, defaults to os.Getenv.
	GetenvFunc func(string) string
}

// DefaultCacheRoot returns the default cache root: ~/.pi/cache
func DefaultCacheRoot() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}
	return filepath.Join(home, ".pi", "cache"), nil
}

// PackagePath returns the expected cache path for a GitHub package version.
func (c *Cache) PackagePath(org, repo, version string) string {
	return filepath.Join(c.Root, "github", org, repo, version)
}

// IsMutableRef returns true if the version string refers to a branch or HEAD
// rather than a fixed tag.
func IsMutableRef(version string) bool {
	switch version {
	case "main", "master", "HEAD":
		return true
	}
	return false
}

func mutableCacheVersion(version string) string {
	return version + "~" + time.Now().Format("20060102")
}

// Fetch ensures the package at org/repo@version is cached. On cache hit, returns
// the cached path immediately with no network call. On cache miss, clones the
// repo and stores it atomically. Returns the path to the cached package root.
func (c *Cache) Fetch(org, repo, version string) (string, error) {
	effectiveVersion := version
	isMutable := IsMutableRef(version)

	if isMutable {
		effectiveVersion = mutableCacheVersion(version)
		if c.WarnWriter != nil {
			fmt.Fprintf(c.WarnWriter,
				"warning: using mutable ref @%s — result may not be reproducible. Pin to a version tag for stability.\n",
				version)
		}
	}

	target := c.PackagePath(org, repo, effectiveVersion)

	// Cache hit
	if info, err := os.Stat(target); err == nil && info.IsDir() {
		if err := c.checkPackageYAML(target); err != nil {
			return "", err
		}
		return target, nil
	}

	// Cache miss — clone and cache
	if err := c.cloneAndCache(org, repo, version, target); err != nil {
		return "", err
	}

	if err := c.checkPackageYAML(target); err != nil {
		os.RemoveAll(target)
		return "", err
	}

	return target, nil
}

// cloneAndCache clones the repo at the given version into a temp dir, then
// atomically renames it to the target path. No partial entry is left on failure.
func (c *Cache) cloneAndCache(org, repo, version, target string) error {
	parentDir := filepath.Dir(target)
	if err := os.MkdirAll(parentDir, 0o755); err != nil {
		return fmt.Errorf("creating cache directory: %w", err)
	}

	tmpDir, err := c.cloneRepo(org, repo, parentDir)
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir) // clean up if rename fails

	if err := c.checkoutVersion(version, tmpDir, org, repo); err != nil {
		return err
	}

	// Remove .git — we only need the working tree
	if err := os.RemoveAll(filepath.Join(tmpDir, ".git")); err != nil {
		return fmt.Errorf("removing .git directory: %w", err)
	}

	if err := os.Rename(tmpDir, target); err != nil {
		return fmt.Errorf("moving package to cache: %w", err)
	}

	return nil
}

// cloneRepo tries SSH first, then HTTPS with GITHUB_TOKEN, then plain HTTPS.
// On success, returns the path to the cloned temp directory.
func (c *Cache) cloneRepo(org, repo, parentDir string) (string, error) {
	sshURL := fmt.Sprintf("git@github.com:%s/%s.git", org, repo)
	httpsURL := fmt.Sprintf("https://github.com/%s/%s.git", org, repo)

	urls := []string{sshURL}
	if token := c.getenv("GITHUB_TOKEN"); token != "" {
		urls = append(urls, fmt.Sprintf("https://%s@github.com/%s/%s.git", token, org, repo))
	}
	urls = append(urls, httpsURL)

	for _, url := range urls {
		tmpDir, err := os.MkdirTemp(parentDir, ".fetch-*")
		if err != nil {
			return "", fmt.Errorf("creating temp directory: %w", err)
		}

		_, _, gitErr := c.git([]string{"clone", "--quiet", url, tmpDir}, "")
		if gitErr == nil {
			return tmpDir, nil
		}

		os.RemoveAll(tmpDir)
	}

	return "", fmt.Errorf("could not fetch %s/%s: check network and that the repo exists.\n"+
		"For private repos:\n"+
		"  • Ensure an SSH key is configured (git@github.com:%s/%s.git)\n"+
		"  • Or set GITHUB_TOKEN env var for HTTPS auth",
		org, repo, org, repo)
}

// checkoutVersion checks out a specific tag/ref in the cloned repo.
func (c *Cache) checkoutVersion(version, repoDir, org, repo string) error {
	_, stderr, err := c.git([]string{"checkout", "--quiet", version}, repoDir)
	if err != nil {
		return fmt.Errorf("could not find version %q for %s/%s: %s",
			version, org, repo, strings.TrimSpace(stderr))
	}
	return nil
}

func (c *Cache) git(args []string, dir string) (string, string, error) {
	if c.GitFunc != nil {
		return c.GitFunc(args, dir)
	}
	return execGit(args, dir)
}

func (c *Cache) getenv(key string) string {
	if c.GetenvFunc != nil {
		return c.GetenvFunc(key)
	}
	return os.Getenv(key)
}

func execGit(args []string, dir string) (string, string, error) {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}
