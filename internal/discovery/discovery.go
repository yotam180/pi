package discovery

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/vyper-tooling/pi/internal/automation"
	"github.com/vyper-tooling/pi/internal/refparser"
)

const (
	PiDir              = ".pi"
	automationFileName = "automation.yaml"
	yamlExt            = ".yaml"
)

const BuiltinPrefix = "pi:"

// PackageInfo describes a resolved package source for tracking provenance.
type PackageInfo struct {
	Source string // original source string from pi.yaml (e.g. "yotam180/pi-common@v1.2")
	Alias  string // alias if declared (e.g. "mytools"), or empty
	Path   string // resolved filesystem path to the package .pi/ directory
}

// OnDemandFetchFunc is called when a GitHub ref is encountered that isn't in
// any declared package. It should fetch the package, merge its automations
// into the result, and return the resolved automation (or an error).
// The source string is "org/repo@version". The ref contains parsed fields.
type OnDemandFetchFunc func(r *Result, ref refparser.AutomationRef) (*automation.Automation, error)

// Result holds all discovered automations and provides lookup.
type Result struct {
	Automations map[string]*automation.Automation
	Builtins    map[string]*automation.Automation // keyed without "pi:" prefix
	names       []string                          // sorted local names
	builtinSet  map[string]bool                   // tracks which names came from builtins
	packageSet  map[string]string                 // automation name → package source
	packages    []PackageInfo                     // resolved packages (ordered)
	aliasMap    map[string]string                 // alias → package source
	packageAuto map[string]map[string]*automation.Automation // source → name → automation

	// OnDemandFetch is called when findInPackage encounters a GitHub ref
	// whose package hasn't been declared or fetched. If nil, an error is returned.
	OnDemandFetch OnDemandFetchFunc
}

// Discover walks the given .pi/ directory, finds all automation YAML files,
// parses them, and returns a Result. It handles two resolution forms:
//   - .pi/docker/up.yaml          → name "docker/up"
//   - .pi/setup/cursor/automation.yaml → name "setup/cursor"
//
// The name: field in automation YAML is optional. When absent, PI derives the
// name from the file path. When present but mismatching the derived name, a
// warning is printed to warnWriter (if non-nil).
//
// Returns an error if two files resolve to the same automation name.
func Discover(piDir string, warnWriter io.Writer) (*Result, error) {
	info, err := os.Stat(piDir)
	if err != nil {
		if os.IsNotExist(err) {
			return &Result{
				Automations: make(map[string]*automation.Automation),
				builtinSet:  make(map[string]bool),
				packageSet:  make(map[string]string),
				aliasMap:    make(map[string]string),
				packageAuto: make(map[string]map[string]*automation.Automation),
			}, nil
		}
		return nil, fmt.Errorf("accessing %s: %w", piDir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", piDir)
	}

	automations := make(map[string]*automation.Automation)
	sources := make(map[string]string) // name → file path (for collision error messages)

	err = filepath.Walk(piDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != yamlExt {
			return nil
		}

		name, ok := deriveName(piDir, path)
		if !ok {
			return nil
		}

		name = normalizeName(name)

		if existingPath, exists := sources[name]; exists {
			return fmt.Errorf("automation name collision: %q resolves from both %s and %s", name, existingPath, path)
		}

		a, err := automation.Load(path)
		if err != nil {
			return fmt.Errorf("loading %s: %w", path, err)
		}

		reconcileAutomationName(a, name, path, warnWriter)

		automations[name] = a
		sources[name] = path
		return nil
	})
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(automations))
	for n := range automations {
		names = append(names, n)
	}
	sort.Strings(names)

	return &Result{
		Automations: automations,
		names:       names,
		builtinSet:  make(map[string]bool),
		packageSet:  make(map[string]string),
		aliasMap:    make(map[string]string),
		packageAuto: make(map[string]map[string]*automation.Automation),
	}, nil
}

// reconcileAutomationName sets the automation's Name from the derived path
// when absent, or emits a warning when the declared name mismatches.
func reconcileAutomationName(a *automation.Automation, derivedName, path string, warnWriter io.Writer) {
	if a.Name == "" {
		a.Name = derivedName
		return
	}
	declared := normalizeName(a.Name)
	if declared != derivedName && warnWriter != nil {
		fmt.Fprintf(warnWriter, "warning: %s: name %q does not match path-derived name %q — the name: field can be removed\n",
			path, a.Name, derivedName)
	}
}

// NewResult creates a Result from a pre-built map and sorted name list.
func NewResult(automations map[string]*automation.Automation, names []string) *Result {
	return &Result{
		Automations: automations,
		names:       names,
		builtinSet:  make(map[string]bool),
		packageSet:  make(map[string]string),
		aliasMap:    make(map[string]string),
		packageAuto: make(map[string]map[string]*automation.Automation),
	}
}

// MergePackage discovers automations from a package directory and merges them.
// The source is the original declaration (e.g. "yotam180/pi-common@v1.2"),
// alias is the optional as: name, and pkgDir is the resolved path to the
// package root (which should contain a .pi/ subdirectory).
// Local automations always take precedence over package automations.
func (r *Result) MergePackage(source, alias, pkgDir string, warnWriter io.Writer) error {
	piDir := filepath.Join(pkgDir, PiDir)
	pkgResult, err := Discover(piDir, warnWriter)
	if err != nil {
		return fmt.Errorf("discovering package %s: %w", source, err)
	}

	if r.packageAuto == nil {
		r.packageAuto = make(map[string]map[string]*automation.Automation)
	}
	r.packageAuto[source] = pkgResult.Automations

	if alias != "" {
		if r.aliasMap == nil {
			r.aliasMap = make(map[string]string)
		}
		r.aliasMap[alias] = source
	}

	r.packages = append(r.packages, PackageInfo{
		Source: source,
		Alias:  alias,
		Path:   pkgDir,
	})

	for name, a := range pkgResult.Automations {
		if _, exists := r.Automations[name]; exists {
			if warnWriter != nil && !r.builtinSet[name] {
				fmt.Fprintf(warnWriter, "warning: package %s automation %q shadowed by local automation\n", source, name)
			}
			continue
		}
		r.Automations[name] = a
		if r.packageSet == nil {
			r.packageSet = make(map[string]string)
		}
		r.packageSet[name] = source
	}

	r.rebuildNames()
	return nil
}

// IsPackage returns true if the given name was provided by an external package.
func (r *Result) IsPackage(name string) bool {
	_, ok := r.packageSet[name]
	return ok
}

// PackageSource returns the package source for a given automation name, or empty string.
func (r *Result) PackageSource(name string) string {
	return r.packageSet[name]
}

// Packages returns the list of merged package infos.
func (r *Result) Packages() []PackageInfo {
	out := make([]PackageInfo, len(r.packages))
	copy(out, r.packages)
	return out
}

// PackageAutomations returns the automation map for a specific package source,
// or nil if the source is not known.
func (r *Result) PackageAutomations(source string) map[string]*automation.Automation {
	return r.packageAuto[source]
}

// KnownAliases returns the alias→source map for package resolution.
func (r *Result) KnownAliases() map[string]bool {
	aliases := make(map[string]bool, len(r.aliasMap))
	for alias := range r.aliasMap {
		aliases[alias] = true
	}
	return aliases
}

// MergeBuiltins incorporates built-in automations into this result.
// Built-ins are stored separately: Find("pi:hello") always resolves to the
// built-in, while Find("hello") resolves to local first, falling back to built-in.
// pi list shows built-ins that don't collide with local names.
func (r *Result) MergeBuiltins(builtinResult *Result) {
	if builtinResult == nil {
		return
	}

	r.Builtins = builtinResult.Automations

	for name, a := range builtinResult.Automations {
		if _, exists := r.Automations[name]; !exists {
			r.Automations[name] = a
			r.builtinSet[name] = true
		}
	}

	r.rebuildNames()
}

// deriveName converts a file path within the .pi/ directory into an automation name.
// Returns the name and true if the file is a valid automation, or ("", false) if
// it should be skipped (e.g., non-automation yaml files).
func deriveName(piDir, path string) (string, bool) {
	rel, err := filepath.Rel(piDir, path)
	if err != nil {
		return "", false
	}

	rel = filepath.ToSlash(rel)
	base := filepath.Base(rel)

	if base == automationFileName {
		// .pi/setup/cursor/automation.yaml → "setup/cursor"
		dir := filepath.Dir(rel)
		if dir == "." {
			return "", false
		}
		return filepath.ToSlash(dir), true
	}

	if filepath.Ext(base) == yamlExt {
		// .pi/docker/up.yaml → "docker/up"
		return strings.TrimSuffix(rel, yamlExt), true
	}

	return "", false
}

// normalizeName cleans up an automation name: lowercase, no leading/trailing slashes.
func normalizeName(name string) string {
	name = strings.ToLower(name)
	name = strings.Trim(name, "/")
	return filepath.ToSlash(name)
}

// Find looks up an automation by name. It uses refparser to classify the
// reference string and dispatches to the appropriate resolution strategy:
//   - RefLocal: checks local automations, falls back to built-ins
//   - RefBuiltin: resolves exclusively from built-ins
//   - RefGitHub, RefFile, RefAlias: not yet supported (returns clear error)
func (r *Result) Find(name string) (*automation.Automation, error) {
	return r.FindWithAliases(name, nil)
}

// FindWithAliases is like Find but accepts known aliases for resolving alias references.
// When knownAliases is nil, the Result's own alias map is used.
func (r *Result) FindWithAliases(name string, knownAliases map[string]bool) (*automation.Automation, error) {
	if knownAliases == nil {
		knownAliases = r.KnownAliases()
	}

	ref, err := refparser.Parse(name, knownAliases)
	if err != nil {
		return nil, fmt.Errorf("invalid automation reference: %w", err)
	}

	switch ref.Type {
	case refparser.RefBuiltin:
		return r.findBuiltin(ref)
	case refparser.RefLocal:
		return r.findLocal(ref)
	case refparser.RefGitHub:
		return r.findInPackage(ref)
	case refparser.RefFile:
		return r.findInPackage(ref)
	case refparser.RefAlias:
		return r.findAlias(ref)
	default:
		return nil, fmt.Errorf("unknown reference type for %q", name)
	}
}

func (r *Result) findBuiltin(ref refparser.AutomationRef) (*automation.Automation, error) {
	if r.Builtins != nil {
		if a, ok := r.Builtins[ref.Path]; ok {
			return a, nil
		}
	}
	return nil, fmt.Errorf("built-in automation %q not found", ref.Raw)
}

func (r *Result) findAlias(ref refparser.AutomationRef) (*automation.Automation, error) {
	source, ok := r.aliasMap[ref.Alias]
	if !ok {
		return nil, fmt.Errorf("alias %q is not declared in packages:", ref.Alias)
	}
	autos, ok := r.packageAuto[source]
	if !ok {
		return nil, fmt.Errorf("package %s for alias %q has not been fetched — run pi setup", source, ref.Alias)
	}
	if a, ok := autos[ref.Path]; ok {
		return a, nil
	}
	return nil, fmt.Errorf("automation %q not found in package %s (alias %q)", ref.Path, source, ref.Alias)
}

func (r *Result) findInPackage(ref refparser.AutomationRef) (*automation.Automation, error) {
	var source string
	if ref.Type == refparser.RefGitHub {
		source = ref.Org + "/" + ref.Repo + "@" + ref.Version
	} else {
		source = "file:" + ref.FSPath
	}

	autos, ok := r.packageAuto[source]
	if !ok {
		// file: sources must be declared — no on-demand fetch
		if ref.Type == refparser.RefFile {
			return nil, fmt.Errorf("package %s not found — add it to packages: in pi.yaml and run pi setup", source)
		}

		// GitHub ref not in declared packages — try on-demand fetch
		if r.OnDemandFetch != nil {
			return r.OnDemandFetch(r, ref)
		}

		return nil, fmt.Errorf("package %s has not been fetched — add it to packages: in pi.yaml and run pi setup", source)
	}

	path := ref.Path
	if path == "" {
		return nil, fmt.Errorf("no automation path specified in reference %q", ref.Raw)
	}

	if a, ok := autos[path]; ok {
		return a, nil
	}
	return nil, fmt.Errorf("automation %q not found in package %s", path, source)
}

func (r *Result) findLocal(ref refparser.AutomationRef) (*automation.Automation, error) {
	if a, ok := r.Automations[ref.Path]; ok {
		return a, nil
	}

	if len(r.names) == 0 {
		return nil, fmt.Errorf("automation %q not found (no automations discovered)", ref.Raw)
	}

	return nil, fmt.Errorf("automation %q not found\n\nAvailable automations:\n%s", ref.Raw, r.formatAvailable())
}

// Names returns a sorted list of all automation names (local + built-in).
func (r *Result) Names() []string {
	out := make([]string, len(r.names))
	copy(out, r.names)
	return out
}

// IsBuiltin returns true if the given name was provided by a built-in automation
// (and not shadowed by a local automation).
func (r *Result) IsBuiltin(name string) bool {
	return r.builtinSet[name]
}

func (r *Result) rebuildNames() {
	r.names = make([]string, 0, len(r.Automations))
	for n := range r.Automations {
		r.names = append(r.names, n)
	}
	sort.Strings(r.names)
}

func (r *Result) formatAvailable() string {
	var b strings.Builder
	for _, name := range r.names {
		a := r.Automations[name]
		if a.Description != "" {
			fmt.Fprintf(&b, "  %-30s %s\n", name, a.Description)
		} else {
			fmt.Fprintf(&b, "  %s\n", name)
		}
	}
	return b.String()
}
