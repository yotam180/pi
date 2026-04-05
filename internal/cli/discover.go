package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/vyper-tooling/pi/internal/automation"
	"github.com/vyper-tooling/pi/internal/builtins"
	"github.com/vyper-tooling/pi/internal/cache"
	"github.com/vyper-tooling/pi/internal/config"
	"github.com/vyper-tooling/pi/internal/discovery"
	"github.com/vyper-tooling/pi/internal/display"
	"github.com/vyper-tooling/pi/internal/refparser"
)

// discoverAll discovers local automations from .pi/ and merges built-in
// automations. Local automations take precedence over built-ins with the
// same name.
func discoverAll(root string) (*discovery.Result, error) {
	return discoverAllWithConfig(root, nil, nil)
}

// discoverAllWithConfig discovers local + built-in + package automations.
// When cfg is non-nil and has packages, they are fetched/verified and merged.
// stderr is used for package fetch status output; nil suppresses status.
func discoverAllWithConfig(root string, cfg *config.ProjectConfig, stderr io.Writer) (*discovery.Result, error) {
	piDir := filepath.Join(root, discovery.PiDir)
	result, err := discovery.Discover(piDir, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("discovering automations: %w", err)
	}

	if cfg != nil && len(cfg.Packages) > 0 {
		if err := mergePackages(result, cfg, root, stderr); err != nil {
			return nil, err
		}
	}

	builtinResult, err := builtins.Discover()
	if err != nil {
		return nil, fmt.Errorf("discovering built-in automations: %w", err)
	}

	result.MergeBuiltins(builtinResult)

	result.OnDemandFetch = newOnDemandFetcher(os.Stderr)

	return result, nil
}

// newOnDemandFetcher returns an OnDemandFetchFunc that fetches GitHub packages
// on first encounter and prints an advisory to stderr. The advisory is shown
// once per package per invocation (deduped via a closure-scoped map).
func newOnDemandFetcher(stderr io.Writer) discovery.OnDemandFetchFunc {
	fetched := make(map[string]bool) // tracks sources already fetched this invocation

	return func(r *discovery.Result, ref refparser.AutomationRef) (*automation.Automation, error) {
		source := ref.Org + "/" + ref.Repo + "@" + ref.Version

		if !fetched[source] {
			cacheRoot, err := cache.DefaultCacheRoot()
			if err != nil {
				return nil, err
			}

			c := &cache.Cache{
				Root:       cacheRoot,
				WarnWriter: stderr,
				PIVersion:  version,
			}

			cachePath := c.PackagePath(ref.Org, ref.Repo, ref.Version)
			wasCached := false
			if info, statErr := os.Stat(cachePath); statErr == nil && info.IsDir() {
				wasCached = true
			}

			pkgDir, err := c.Fetch(ref.Org, ref.Repo, ref.Version)
			if err != nil {
				return nil, fmt.Errorf("on-demand fetch of %s failed: %w", source, err)
			}

			if err := r.MergePackage(source, "", pkgDir, stderr); err != nil {
				return nil, err
			}

			fetched[source] = true

			// Print advisory only when a live network fetch happened (not from cache)
			if !wasCached {
				printOnDemandAdvisory(stderr, source)
			}
		}

		// Package is now merged — look up the automation
		path := ref.Path
		if path == "" {
			return nil, fmt.Errorf("no automation path specified in reference %q", ref.Raw)
		}

		if a, ok := r.Automations[path]; ok {
			return a, nil
		}

		autos := r.PackageAutomations(source)
		if a, ok := autos[path]; ok {
			return a, nil
		}

		return nil, fmt.Errorf("automation %q not found in package %s", path, source)
	}
}

// printOnDemandAdvisory prints a user-facing advisory to stderr when a package
// was fetched on demand (not from cache). Nudges the user to declare it.
func printOnDemandAdvisory(w io.Writer, source string) {
	if w == nil {
		return
	}
	printer := display.NewForWriter(w)
	printer.PackageFetch("↓", source, "fetched (on demand)", "")
	fmt.Fprintf(w, "\n  tip: add to pi.yaml to avoid fetching on every fresh clone:\n\n")
	fmt.Fprintf(w, "    packages:\n")
	fmt.Fprintf(w, "      - %s\n\n", source)
}

// mergePackages fetches/verifies each declared package and merges its
// automations into the result. root is the project root for resolving
// relative file: paths.
func mergePackages(result *discovery.Result, cfg *config.ProjectConfig, root string, stderr io.Writer) error {
	var printer *display.Printer
	if stderr != nil {
		printer = display.NewForWriter(stderr)
	}

	for _, pkg := range cfg.Packages {
		pkgDir, err := resolvePackageSource(pkg, root, stderr, printer)
		if err != nil {
			return fmt.Errorf("package %s: %w", pkg.Source, err)
		}
		if pkgDir == "" {
			continue // file: source not found — warning printed, non-fatal
		}

		if err := result.MergePackage(pkg.Source, pkg.As, pkgDir, os.Stderr); err != nil {
			return err
		}
	}
	return nil
}

// resolvePackageSource fetches a GitHub package or verifies a file: source.
// Returns the path to the package root, or empty string if a file: source
// is missing (warning printed, non-fatal).
func resolvePackageSource(pkg config.PackageEntry, root string, stderr io.Writer, printer *display.Printer) (string, error) {
	if pkg.IsFileSource() {
		return resolveFilePackage(pkg, root, stderr, printer)
	}
	return resolveGitHubPackage(pkg, stderr, printer)
}

func resolveFilePackage(pkg config.PackageEntry, root string, stderr io.Writer, printer *display.Printer) (string, error) {
	fsPath := pkg.FilePath()

	if !filepath.IsAbs(fsPath) {
		fsPath = filepath.Join(root, fsPath)
	}

	info, err := os.Stat(fsPath)
	if err != nil || !info.IsDir() {
		if printer != nil {
			detail := ""
			if pkg.As != "" {
				detail = "alias: " + pkg.As
			}
			printer.PackageFetch("⚠", pkg.Source, "not found", detail)
		}
		if stderr != nil {
			fmt.Fprintf(stderr, "warning: package %s path does not exist: %s\n", pkg.Source, fsPath)
		}
		return "", nil // non-fatal
	}

	if printer != nil {
		detail := ""
		if pkg.As != "" {
			detail = "alias: " + pkg.As
		}
		printer.PackageFetch("✓", pkg.Source, "found", detail)
	}
	return fsPath, nil
}

func resolveGitHubPackage(pkg config.PackageEntry, stderr io.Writer, printer *display.Printer) (string, error) {
	ref, err := refparser.Parse(pkg.Source, nil)
	if err != nil {
		return "", fmt.Errorf("invalid package source %q: %w", pkg.Source, err)
	}
	if ref.Type != refparser.RefGitHub {
		return "", fmt.Errorf("invalid package source %q: expected org/repo@version format", pkg.Source)
	}

	cacheRoot, err := cache.DefaultCacheRoot()
	if err != nil {
		return "", err
	}

	c := &cache.Cache{
		Root:       cacheRoot,
		WarnWriter: stderr,
		PIVersion:  version,
	}

	cachePath := c.PackagePath(ref.Org, ref.Repo, ref.Version)
	if info, statErr := os.Stat(cachePath); statErr == nil && info.IsDir() {
		if printer != nil {
			printer.PackageFetch("✓", pkg.Source, "cached", "")
		}
		return cachePath, nil
	}

	if printer != nil {
		printer.PackageFetch("↓", pkg.Source, "fetching...", "")
	}

	path, err := c.Fetch(ref.Org, ref.Repo, ref.Version)
	if err != nil {
		if printer != nil {
			printer.PackageFetch("✗", pkg.Source, "failed", "")
		}
		return "", fmt.Errorf("fetching %s: %w", pkg.Source, err)
	}

	if printer != nil {
		printer.PackageFetch("✓", pkg.Source, "cached", "")
	}
	return path, nil
}
