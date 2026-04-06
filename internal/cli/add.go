package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vyper-tooling/pi/internal/config"
	"github.com/vyper-tooling/pi/internal/display"
	"github.com/vyper-tooling/pi/internal/refparser"
)

func newAddCmd() *cobra.Command {
	var asFlag string

	cmd := &cobra.Command{
		Use:   "add <source>",
		Short: "Add a package dependency to pi.yaml",
		Long: `Add a package dependency to pi.yaml and fetch it immediately (for GitHub sources).

Examples:
  pi add yotam180/pi-common@v1.2
  pi add file:~/shared-automations
  pi add file:~/my-automations --as mytools`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			startDir, err := getwd()
			if err != nil {
				return err
			}
			return runAdd(startDir, args[0], asFlag, os.Stdout, os.Stderr)
		},
	}

	cmd.Flags().StringVar(&asFlag, "as", "", "alias for the package (used in automation references)")

	return cmd
}

// runAdd implements the pi add logic. Extracted for testability.
func runAdd(startDir, source, alias string, stdout, stderr io.Writer) error {
	return runAddWithFetcher(startDir, source, alias, stdout, stderr, nil)
}

// runAddWithFetcher is the internal implementation of runAdd that accepts an
// optional PackageFetcher for testability. When fetcher is nil, a
// CachePackageFetcher is created for GitHub sources.
func runAddWithFetcher(startDir, source, alias string, stdout, stderr io.Writer, fetcher PackageFetcher) error {
	pc, err := resolveProject(startDir)
	if err != nil {
		return err
	}
	root := pc.Root

	entry := config.PackageEntry{
		Source: source,
		As:     alias,
	}

	ref, err := refparser.Parse(source, nil)
	if err != nil {
		return err
	}

	printer := display.NewForWriter(stderr)

	switch ref.Type {
	case refparser.RefGitHub:
		f := fetcher
		if f == nil {
			f, err = NewCachePackageFetcher(stderr)
			if err != nil {
				return err
			}
		}
		if err := fetchGitHubPackage(ref, stderr, printer, f); err != nil {
			return err
		}

	case refparser.RefFile:
		// file: sources don't need fetching, just validation that the path is reasonable

	case refparser.RefLocal:
		if err := validateGitHubSource(source); err != nil {
			return err
		}
		return fmt.Errorf("invalid package source %q: must be org/repo@version or file:<path>", source)

	default:
		return fmt.Errorf("invalid package source %q: must be org/repo@version or file:<path>", source)
	}

	if err := config.AddPackage(root, entry); err != nil {
		var dup *config.DuplicatePackageError
		if errors.As(err, &dup) {
			printer.Plain("  ✓  %s already in %s\n", dup.Source, config.FileName)
			return nil
		}
		return err
	}

	printer.Plain("  ✓  added %s to %s\n", source, config.FileName)
	return nil
}

// fetchGitHubPackage validates and fetches a GitHub package into the cache.
func fetchGitHubPackage(ref refparser.AutomationRef, stderr io.Writer, printer *display.Printer, fetcher PackageFetcher) error {
	if ref.Version == "" {
		return fmt.Errorf("version required — use %s/%s@<tag>", ref.Org, ref.Repo)
	}

	source := ref.Org + "/" + ref.Repo + "@" + ref.Version
	printer.PackageFetch(display.StatusInProgress, source, "fetching...", "")

	path, wasCached, err := fetcher.Fetch(ref.Org, ref.Repo, ref.Version)
	if err != nil {
		printer.PackageFetch(display.StatusFailed, source, "failed", "")
		return fmt.Errorf("fetching %s: %w", source, err)
	}

	_ = path // path is used by callers that merge; add.go only needs the side effect

	status := "cached"
	if !wasCached {
		status = "fetched"
	}
	printer.PackageFetch(display.StatusSuccessCached, source, status, "")
	return nil
}

// validateGitHubSource checks that a GitHub source has a version tag.
func validateGitHubSource(source string) error {
	if !strings.Contains(source, "@") {
		parts := strings.SplitN(source, "/", 2)
		if len(parts) == 2 {
			return fmt.Errorf("version required — use pi add %s@<tag>", source)
		}
	}
	return nil
}
