package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// SetupEntry is re-exported here for writer use — same type as config.SetupEntry.

// AddSetupEntry adds a setup entry to pi.yaml in the given directory.
// It reads the file, checks for duplicates, and appends the entry to the
// setup: block (creating it if absent). Preserves existing file content.
func AddSetupEntry(dir string, entry SetupEntry) error {
	path := filepath.Join(dir, FileName)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s not found in %s", FileName, dir)
		}
		return fmt.Errorf("reading %s: %w", path, err)
	}

	cfg, err := Load(dir)
	if err != nil {
		return err
	}

	if isSetupDuplicate(cfg.Setup, entry) {
		return &DuplicateSetupEntryError{Run: entry.Run}
	}

	content := string(data)
	entryYAML := FormatSetupEntry(entry)
	updated, err := insertSetupEntry(content, entryYAML)
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}

	if _, err := Load(dir); err != nil {
		return fmt.Errorf("validation after update failed: %w", err)
	}

	return nil
}

// DuplicateSetupEntryError is returned when trying to add a setup entry that's
// already declared in pi.yaml.
type DuplicateSetupEntryError struct {
	Run string
}

func (e *DuplicateSetupEntryError) Error() string {
	return fmt.Sprintf("setup entry %q is already declared in %s", e.Run, FileName)
}

// isSetupDuplicate checks if a setup entry with the same run target and
// matching with/if fields already exists.
func isSetupDuplicate(entries []SetupEntry, entry SetupEntry) bool {
	for _, e := range entries {
		if e.Run != entry.Run {
			continue
		}
		if e.If != entry.If {
			continue
		}
		if !mapsEqual(e.With, entry.With) {
			continue
		}
		return true
	}
	return false
}

func mapsEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

// FormatSetupEntry formats a SetupEntry as one or more YAML lines for display/insertion.
func FormatSetupEntry(entry SetupEntry) string {
	needsObjectForm := entry.If != "" || len(entry.With) > 0

	if !needsObjectForm {
		return "  - " + entry.Run
	}

	var sb strings.Builder
	sb.WriteString("  - run: " + entry.Run)

	if entry.If != "" {
		sb.WriteString("\n    if: " + entry.If)
	}

	if len(entry.With) > 0 {
		sb.WriteString("\n    with:")
		keys := sortedKeys(entry.With)
		for _, k := range keys {
			v := entry.With[k]
			sb.WriteString(fmt.Sprintf("\n      %s: %q", k, v))
		}
	}

	return sb.String()
}

// sortedKeys returns the keys of a map in sorted order.
func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// insertSetupEntry appends the rendered YAML entry to the setup: block in the
// file content. Creates the setup: block if it doesn't exist.
func insertSetupEntry(content string, entryYAML string) (string, error) {
	lines := strings.Split(content, "\n")
	setupIdx := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "setup:" {
			setupIdx = i
			break
		}
	}

	if setupIdx == -1 {
		return appendNewSetupBlock(content, entryYAML), nil
	}

	return appendToExistingSetupBlock(lines, setupIdx, entryYAML), nil
}

// appendNewSetupBlock adds a setup: block at the end of the file.
func appendNewSetupBlock(content string, entryYAML string) string {
	trimmed := strings.TrimRight(content, "\n")
	return trimmed + "\n\nsetup:\n" + entryYAML + "\n"
}

// appendToExistingSetupBlock inserts a new entry at the end of the setup: list.
func appendToExistingSetupBlock(lines []string, setupIdx int, entryYAML string) string {
	insertIdx := setupIdx + 1

	for insertIdx < len(lines) {
		line := lines[insertIdx]
		if line == "" {
			insertIdx++
			continue
		}

		trimmed := strings.TrimSpace(line)

		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") && trimmed != "" {
			break
		}

		insertIdx++
	}

	for insertIdx > setupIdx+1 && strings.TrimSpace(lines[insertIdx-1]) == "" {
		insertIdx--
	}

	entryLines := strings.Split(entryYAML, "\n")

	result := make([]string, 0, len(lines)+len(entryLines))
	result = append(result, lines[:insertIdx]...)
	result = append(result, entryLines...)
	result = append(result, lines[insertIdx:]...)

	return strings.Join(result, "\n")
}

// AddPackage adds a package entry to pi.yaml in the given directory.
// It reads the file, checks for duplicates, and appends the entry to the
// packages: block (creating it if absent). Preserves existing file content.
func AddPackage(dir string, entry PackageEntry) error {
	path := filepath.Join(dir, FileName)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s not found in %s", FileName, dir)
		}
		return fmt.Errorf("reading %s: %w", path, err)
	}

	cfg, err := Load(dir)
	if err != nil {
		return err
	}

	if isDuplicate(cfg.Packages, entry) {
		return &DuplicatePackageError{Source: entry.Source}
	}

	content := string(data)
	updated, err := insertPackageEntry(content, entry)
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}

	// Re-validate to ensure the updated file is still valid
	if _, err := Load(dir); err != nil {
		return fmt.Errorf("validation after update failed: %w", err)
	}

	return nil
}

// DuplicatePackageError is returned when trying to add a package that's
// already declared in pi.yaml.
type DuplicatePackageError struct {
	Source string
}

func (e *DuplicatePackageError) Error() string {
	return fmt.Sprintf("package %s is already declared in %s", e.Source, FileName)
}

// isDuplicate checks if a package with the same source is already in the list.
func isDuplicate(packages []PackageEntry, entry PackageEntry) bool {
	for _, p := range packages {
		if p.Source == entry.Source {
			return true
		}
	}
	return false
}

// insertPackageEntry appends the package entry to the packages: block in the
// YAML content. Creates the packages: block if it doesn't exist.
func insertPackageEntry(content string, entry PackageEntry) (string, error) {
	entryYAML := formatPackageEntry(entry)

	lines := strings.Split(content, "\n")
	packagesIdx := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "packages:" {
			packagesIdx = i
			break
		}
	}

	if packagesIdx == -1 {
		return appendNewPackagesBlock(content, entryYAML), nil
	}

	return appendToExistingPackagesBlock(lines, packagesIdx, entryYAML), nil
}

// formatPackageEntry formats a PackageEntry as a YAML list item string.
// Simple sources get the short form; sources with aliases get the object form.
func formatPackageEntry(entry PackageEntry) string {
	if entry.As == "" {
		return "  - " + entry.Source
	}
	return "  - source: " + entry.Source + "\n    as: " + entry.As
}

// appendNewPackagesBlock adds a packages: block at the end of the file.
func appendNewPackagesBlock(content string, entryYAML string) string {
	trimmed := strings.TrimRight(content, "\n")
	return trimmed + "\n\npackages:\n" + entryYAML + "\n"
}

// appendToExistingPackagesBlock inserts a new entry at the end of the
// packages: list items.
func appendToExistingPackagesBlock(lines []string, packagesIdx int, entryYAML string) string {
	insertIdx := packagesIdx + 1

	for insertIdx < len(lines) {
		line := lines[insertIdx]
		if line == "" {
			insertIdx++
			continue
		}

		trimmed := strings.TrimSpace(line)

		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") && trimmed != "" {
			break
		}

		insertIdx++
	}

	// Walk back over trailing blank lines within the packages block
	for insertIdx > packagesIdx+1 && strings.TrimSpace(lines[insertIdx-1]) == "" {
		insertIdx--
	}

	entryLines := strings.Split(entryYAML, "\n")

	result := make([]string, 0, len(lines)+len(entryLines))
	result = append(result, lines[:insertIdx]...)
	result = append(result, entryLines...)
	result = append(result, lines[insertIdx:]...)

	return strings.Join(result, "\n")
}
