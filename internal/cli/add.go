package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vyper-tooling/pi/internal/cache"
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
		if err := fetchGitHubPackage(ref, stderr, printer); err != nil {
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
		if dup, ok := err.(*config.DuplicatePackageError); ok {
			printer.Plain("  ✓  %s already in %s\n", dup.Source, config.FileName)
			return nil
		}
		return err
	}

	printer.Plain("  ✓  added %s to %s\n", source, config.FileName)
	return nil
}

// fetchGitHubPackage validates and fetches a GitHub package into the cache.
func fetchGitHubPackage(ref refparser.AutomationRef, stderr io.Writer, printer *display.Printer) error {
	if ref.Version == "" {
		return fmt.Errorf("version required — use %s/%s@<tag>", ref.Org, ref.Repo)
	}

	source := ref.Org + "/" + ref.Repo + "@" + ref.Version

	cacheRoot, err := cache.DefaultCacheRoot()
	if err != nil {
		return err
	}

	c := &cache.Cache{
		Root:       cacheRoot,
		WarnWriter: stderr,
		PIVersion:  version,
	}

	cachePath := c.PackagePath(ref.Org, ref.Repo, ref.Version)
	if info, statErr := os.Stat(cachePath); statErr == nil && info.IsDir() {
		printer.PackageFetch("✓", source, "cached", "")
		return nil
	}

	printer.PackageFetch("↓", source, "fetching...", "")

	if _, err := c.Fetch(ref.Org, ref.Repo, ref.Version); err != nil {
		printer.PackageFetch("✗", source, "failed", "")
		return fmt.Errorf("fetching %s: %w", source, err)
	}

	printer.PackageFetch("✓", source, "cached", "")
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
