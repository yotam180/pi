package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/vyper-tooling/pi/internal/automation"
	"github.com/vyper-tooling/pi/internal/conditions"
	"github.com/vyper-tooling/pi/internal/config"
	"github.com/vyper-tooling/pi/internal/discovery"
	"github.com/vyper-tooling/pi/internal/executor"
)

// Context holds everything a validation check needs. Checks receive this
// as read-only context — they append errors to Result via the Runner.
type Context struct {
	Root      string
	Config    *config.ProjectConfig
	Discovery *discovery.Result
}

// Result holds the outcome of a project validation.
type Result struct {
	Errors          []string
	Warnings        []string
	AutomationCount int
	ShortcutCount   int
	SetupCount      int
}

// Check performs a single validation concern. Implementations are stateless —
// all state comes from Context.
type Check interface {
	Name() string
	Run(ctx *Context) []string
}

// CheckFunc adapts a plain function to the Check interface.
type CheckFunc struct {
	CheckName string
	Fn        func(ctx *Context) []string
}

func (c CheckFunc) Name() string              { return c.CheckName }
func (c CheckFunc) Run(ctx *Context) []string { return c.Fn(ctx) }

// WarnCheck produces non-fatal warnings (separate from errors).
type WarnCheck interface {
	Name() string
	Run(ctx *Context) []string
}

// WarnCheckFunc adapts a plain function to the WarnCheck interface.
type WarnCheckFunc struct {
	CheckName string
	Fn        func(ctx *Context) []string
}

func (c WarnCheckFunc) Name() string              { return c.CheckName }
func (c WarnCheckFunc) Run(ctx *Context) []string { return c.Fn(ctx) }

// Runner collects checks and runs them against a project context.
type Runner struct {
	checks     []Check
	warnChecks []WarnCheck
}

// NewRunner creates an empty runner.
func NewRunner() *Runner {
	return &Runner{}
}

// Register adds a check to the runner. Checks run in registration order.
func (r *Runner) Register(c Check) {
	r.checks = append(r.checks, c)
}

// RegisterWarn adds a warning check to the runner.
func (r *Runner) RegisterWarn(c WarnCheck) {
	r.warnChecks = append(r.warnChecks, c)
}

// Run executes all registered checks and returns the aggregated result.
// Warning checks only run when includeWarnings is true.
func (r *Runner) Run(ctx *Context) Result {
	return r.RunWithOpts(ctx, false)
}

// RunWithOpts executes checks and optionally includes warning checks.
func (r *Runner) RunWithOpts(ctx *Context, includeWarnings bool) Result {
	result := Result{
		AutomationCount: len(ctx.Discovery.Names()),
		ShortcutCount:   len(ctx.Config.Shortcuts),
		SetupCount:      len(ctx.Config.Setup),
	}
	for _, c := range r.checks {
		result.Errors = append(result.Errors, c.Run(ctx)...)
	}
	if includeWarnings {
		for _, c := range r.warnChecks {
			result.Warnings = append(result.Warnings, c.Run(ctx)...)
		}
	}
	return result
}

// Checks returns the registered check count (useful for tests).
func (r *Runner) Checks() int {
	return len(r.checks)
}

// WarnChecks returns the registered warning check count.
func (r *Runner) WarnChecks() int {
	return len(r.warnChecks)
}

// DefaultRunner returns a Runner pre-loaded with all built-in validation checks.
func DefaultRunner() *Runner {
	r := NewRunner()
	r.Register(CheckFunc{CheckName: "shortcut-refs", Fn: checkShortcutRefs})
	r.Register(CheckFunc{CheckName: "setup-refs", Fn: checkSetupRefs})
	r.Register(CheckFunc{CheckName: "run-step-refs", Fn: checkRunStepRefs})
	r.Register(CheckFunc{CheckName: "file-refs", Fn: checkFileReferences})
	r.Register(CheckFunc{CheckName: "shortcut-inputs", Fn: checkShortcutInputs})
	r.Register(CheckFunc{CheckName: "setup-inputs", Fn: checkSetupInputs})
	r.Register(CheckFunc{CheckName: "run-step-inputs", Fn: checkRunStepInputs})
	r.Register(CheckFunc{CheckName: "circular-deps", Fn: checkCircularDeps})
	r.Register(CheckFunc{CheckName: "conditions", Fn: checkConditions})
	r.Register(CheckFunc{CheckName: "unknown-fields", Fn: checkUnknownFields})
	r.Register(CheckFunc{CheckName: "unknown-pi-yaml-fields", Fn: checkPiYamlUnknownFields})
	r.RegisterWarn(WarnCheckFunc{CheckName: "missing-description", Fn: warnMissingDescription})
	r.RegisterWarn(WarnCheckFunc{CheckName: "unused-automations", Fn: warnUnusedAutomations})
	r.RegisterWarn(WarnCheckFunc{CheckName: "shortcut-shadowing", Fn: warnShortcutShadowing})
	return r
}

// --- Individual checks ---

func checkShortcutRefs(ctx *Context) []string {
	var errs []string
	for name, shortcut := range ctx.Config.Shortcuts {
		if _, err := ctx.Discovery.Find(shortcut.Run); err != nil {
			errs = append(errs,
				fmt.Sprintf("pi.yaml: shortcut %q references unknown automation %q", name, shortcut.Run))
		}
	}
	return errs
}

func checkSetupRefs(ctx *Context) []string {
	var errs []string
	for i, entry := range ctx.Config.Setup {
		if _, err := ctx.Discovery.Find(entry.Run); err != nil {
			errs = append(errs,
				fmt.Sprintf("pi.yaml: setup[%d] references unknown automation %q", i, entry.Run))
		}
	}
	return errs
}

func checkRunStepRefs(ctx *Context) []string {
	var errs []string
	for _, name := range ctx.Discovery.Names() {
		a := ctx.Discovery.Automations[name]
		automation.WalkSteps(a, func(step automation.Step, loc automation.StepLocation) {
			if step.Type != automation.StepTypeRun {
				return
			}
			if _, err := ctx.Discovery.Find(step.Value); err != nil {
				errs = append(errs,
					fmt.Sprintf("%s run: references unknown automation %q", loc.FormatPath(a.Name), step.Value))
			}
		})
	}
	return errs
}

func checkFileReferences(ctx *Context) []string {
	var errs []string
	for _, name := range ctx.Discovery.Names() {
		if ctx.Discovery.IsBuiltin(name) {
			continue
		}
		a := ctx.Discovery.Automations[name]
		automation.WalkSteps(a, func(step automation.Step, loc automation.StepLocation) {
			if step.Type == automation.StepTypeRun {
				return
			}
			if !executor.IsFilePath(step.Value) {
				return
			}
			resolved := filepath.Join(a.Dir(), step.Value)
			if _, err := os.Stat(resolved); err != nil {
				stepType := string(step.Type)
				if loc.IsScalar {
					stepType = "bash"
				}
				errs = append(errs,
					fmt.Sprintf("%s %s: file not found: %s (resolved to %s)", loc.FormatPath(a.Name), stepType, step.Value, resolved))
			}
		})
	}
	return errs
}

func checkShortcutInputs(ctx *Context) []string {
	var errs []string
	for name, shortcut := range ctx.Config.Shortcuts {
		if len(shortcut.With) == 0 {
			continue
		}
		target, err := ctx.Discovery.Find(shortcut.Run)
		if err != nil {
			continue
		}
		for _, msg := range CheckWithInputs(shortcut.With, target) {
			errs = append(errs,
				fmt.Sprintf("pi.yaml: shortcut %q %s", name, msg))
		}
	}
	return errs
}

func checkSetupInputs(ctx *Context) []string {
	var errs []string
	for i, entry := range ctx.Config.Setup {
		if len(entry.With) == 0 {
			continue
		}
		target, err := ctx.Discovery.Find(entry.Run)
		if err != nil {
			continue
		}
		for _, msg := range CheckWithInputs(entry.With, target) {
			errs = append(errs,
				fmt.Sprintf("pi.yaml: setup[%d] %s", i, msg))
		}
	}
	return errs
}

func checkRunStepInputs(ctx *Context) []string {
	var errs []string
	for _, name := range ctx.Discovery.Names() {
		a := ctx.Discovery.Automations[name]
		automation.WalkSteps(a, func(step automation.Step, loc automation.StepLocation) {
			if step.Type != automation.StepTypeRun || len(step.With) == 0 {
				return
			}
			target, err := ctx.Discovery.Find(step.Value)
			if err != nil {
				return
			}
			for _, msg := range CheckWithInputs(step.With, target) {
				errs = append(errs,
					fmt.Sprintf("%s %s", loc.FormatPath(a.Name), msg))
			}
		})
	}
	return errs
}

func checkCircularDeps(ctx *Context) []string {
	graph := BuildRunGraph(ctx.Discovery)
	cycles := DetectCycles(graph)
	var errs []string
	for _, cycle := range cycles {
		errs = append(errs,
			fmt.Sprintf("circular dependency: %s", strings.Join(cycle, " → ")))
	}
	return errs
}

func checkConditions(ctx *Context) []string {
	var errs []string
	for _, name := range ctx.Discovery.Names() {
		a := ctx.Discovery.Automations[name]

		if a.If != "" {
			if err := conditions.ValidateConditionExpr(a.If); err != nil {
				errs = append(errs,
					fmt.Sprintf("%s if: %s", name, err))
			}
		}

		automation.WalkSteps(a, func(step automation.Step, loc automation.StepLocation) {
			if step.If != "" {
				if err := conditions.ValidateConditionExpr(step.If); err != nil {
					errs = append(errs,
						fmt.Sprintf("%s if: %s", loc.FormatPath(a.Name), err))
				}
			}
		})
	}
	return errs
}

// --- Exported helpers (used by cli/validate.go tests and other packages) ---

// CheckWithInputs validates that with: keys match the target automation's
// declared inputs. Returns a list of error messages (empty if all valid).
func CheckWithInputs(with map[string]string, target *automation.Automation) []string {
	if len(with) == 0 {
		return nil
	}
	if len(target.Inputs) == 0 {
		return []string{fmt.Sprintf("passes with: to %q which has no declared inputs", target.Name)}
	}
	var msgs []string
	for key := range with {
		if _, ok := target.Inputs[key]; !ok {
			available := strings.Join(target.InputKeys, ", ")
			msgs = append(msgs, fmt.Sprintf("with: key %q is not a declared input of %q (available: %s)", key, target.Name, available))
		}
	}
	sort.Strings(msgs)
	return msgs
}

// BuildRunGraph extracts a directed graph from all run: steps in discovered automations.
// Keys are automation names; values are deduplicated lists of target automation names.
func BuildRunGraph(disc *discovery.Result) map[string][]string {
	graph := make(map[string][]string)
	for _, name := range disc.Names() {
		a := disc.Automations[name]
		seen := make(map[string]bool)
		automation.WalkSteps(a, func(step automation.Step, loc automation.StepLocation) {
			if step.Type != automation.StepTypeRun {
				return
			}
			target, err := disc.Find(step.Value)
			if err != nil {
				return
			}
			if !seen[target.Name] {
				seen[target.Name] = true
				graph[name] = append(graph[name], target.Name)
			}
		})
		if _, exists := graph[name]; !exists {
			graph[name] = nil
		}
	}
	return graph
}

// DetectCycles finds all unique cycles in a directed graph using DFS.
// Returns each cycle as a slice of names (e.g., ["A", "B", "C", "A"]).
func DetectCycles(graph map[string][]string) [][]string {
	const (
		white = 0
		gray  = 1
		black = 2
	)

	color := make(map[string]int)
	path := make([]string, 0)
	var cycles [][]string
	reported := make(map[string]bool)

	var dfs func(node string)
	dfs = func(node string) {
		color[node] = gray
		path = append(path, node)

		for _, neighbor := range graph[node] {
			if color[neighbor] == gray {
				cycleStart := -1
				for i, n := range path {
					if n == neighbor {
						cycleStart = i
						break
					}
				}
				if cycleStart >= 0 {
					cycle := make([]string, len(path)-cycleStart+1)
					copy(cycle, path[cycleStart:])
					cycle[len(cycle)-1] = neighbor

					key := NormalizeCycleKey(cycle)
					if !reported[key] {
						reported[key] = true
						cycles = append(cycles, cycle)
					}
				}
			} else if color[neighbor] == white {
				dfs(neighbor)
			}
		}

		path = path[:len(path)-1]
		color[node] = black
	}

	nodes := make([]string, 0, len(graph))
	for node := range graph {
		nodes = append(nodes, node)
	}
	sort.Strings(nodes)

	for _, node := range nodes {
		if color[node] == white {
			dfs(node)
		}
	}

	return cycles
}

// NormalizeCycleKey produces a canonical string for a cycle so that the same
// cycle discovered from different starting nodes is only reported once.
func NormalizeCycleKey(cycle []string) string {
	if len(cycle) <= 1 {
		return strings.Join(cycle, "→")
	}
	ring := cycle[:len(cycle)-1]
	minIdx := 0
	for i, name := range ring {
		if name < ring[minIdx] {
			minIdx = i
		}
	}
	rotated := make([]string, len(ring))
	for i := range ring {
		rotated[i] = ring[(minIdx+i)%len(ring)]
	}
	return strings.Join(rotated, "→")
}
