package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vyper-tooling/pi/internal/automation"
	"github.com/vyper-tooling/pi/internal/config"
	"github.com/vyper-tooling/pi/internal/discovery"
	"github.com/vyper-tooling/pi/internal/executor"
)

func newValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate all automation files and config",
		Long: `Statically validate pi.yaml and all automation YAML files in .pi/ without
executing anything. Checks for schema errors, broken references, missing
script files, input mismatches, and configuration mistakes.

Checks performed:
  - pi.yaml parsing and schema validation
  - All .pi/*.yaml automation parsing and schema validation
  - Shortcut targets reference existing automations
  - Setup entry targets reference existing automations
  - run: steps reference existing automations
  - File-path step values (*.sh, *.py, *.ts) reference existing files
  - with: keys on shortcuts match target automation's declared inputs
  - with: keys on setup entries match target automation's declared inputs
  - with: keys on run: steps match target automation's declared inputs

Exits with code 0 if all checks pass, or code 1 if any errors are found.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := getwd()
			if err != nil {
				return err
			}
			return runValidate(cwd, cmd.OutOrStdout(), cmd.ErrOrStderr())
		},
	}
}

// ValidationResult holds the outcome of a project validation.
type ValidationResult struct {
	Errors          []string
	AutomationCount int
	ShortcutCount   int
	SetupCount      int
}

func runValidate(startDir string, stdout, stderr io.Writer) error {
	pc, err := resolveProject(startDir)
	if err != nil {
		return err
	}

	result := validateProject(pc.Root)

	if len(result.Errors) > 0 {
		for _, e := range result.Errors {
			fmt.Fprintf(stderr, "✗ %s\n", e)
		}
		fmt.Fprintf(stderr, "\nValidation failed: %d error(s)\n", len(result.Errors))
		return &executor.ExitError{Code: 1}
	}

	fmt.Fprintf(stdout, "✓ Validated %d automation(s), %d shortcut(s), %d setup entry(ies)\n",
		result.AutomationCount, result.ShortcutCount, result.SetupCount)
	return nil
}

func validateProject(root string) ValidationResult {
	var result ValidationResult

	cfg, cfgErr := config.Load(root)
	if cfgErr != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("pi.yaml: %s", cfgErr))
		return result
	}

	disc, discErr := discoverAllWithConfig(root, cfg, nil)
	if discErr != nil {
		result.Errors = append(result.Errors, fmt.Sprintf(".pi/: %s", discErr))
		return result
	}

	result.AutomationCount = len(disc.Names())
	result.ShortcutCount = len(cfg.Shortcuts)
	result.SetupCount = len(cfg.Setup)

	validateShortcutRefs(cfg, disc, &result)
	validateSetupRefs(cfg, disc, &result)
	validateRunStepRefs(disc, &result)
	validateFileReferences(disc, &result)
	validateShortcutInputs(cfg, disc, &result)
	validateSetupInputs(cfg, disc, &result)
	validateRunStepInputs(disc, &result)

	return result
}

func validateShortcutRefs(cfg *config.ProjectConfig, disc *discovery.Result, result *ValidationResult) {
	for name, shortcut := range cfg.Shortcuts {
		if _, err := disc.Find(shortcut.Run); err != nil {
			result.Errors = append(result.Errors,
				fmt.Sprintf("pi.yaml: shortcut %q references unknown automation %q", name, shortcut.Run))
		}
	}
}

func validateSetupRefs(cfg *config.ProjectConfig, disc *discovery.Result, result *ValidationResult) {
	for i, entry := range cfg.Setup {
		if _, err := disc.Find(entry.Run); err != nil {
			result.Errors = append(result.Errors,
				fmt.Sprintf("pi.yaml: setup[%d] references unknown automation %q", i, entry.Run))
		}
	}
}

func validateRunStepRefs(disc *discovery.Result, result *ValidationResult) {
	for _, name := range disc.Names() {
		a := disc.Automations[name]
		automation.WalkSteps(a, func(step automation.Step, loc automation.StepLocation) {
			if step.Type != automation.StepTypeRun {
				return
			}
			if _, err := disc.Find(step.Value); err != nil {
				result.Errors = append(result.Errors,
					fmt.Sprintf("%s run: references unknown automation %q", loc.FormatPath(a.Name), step.Value))
			}
		})
	}
}

// validateFileReferences checks that file-path step values reference files
// that actually exist on disk. Built-in automations are skipped since they
// use inline scripts only and have no real filesystem directory.
func validateFileReferences(disc *discovery.Result, result *ValidationResult) {
	for _, name := range disc.Names() {
		if disc.IsBuiltin(name) {
			continue
		}
		a := disc.Automations[name]
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
				result.Errors = append(result.Errors,
					fmt.Sprintf("%s %s: file not found: %s (resolved to %s)", loc.FormatPath(a.Name), stepType, step.Value, resolved))
			}
		})
	}
}

// validateShortcutInputs checks that with: keys on shortcuts match declared
// inputs on the target automation.
func validateShortcutInputs(cfg *config.ProjectConfig, disc *discovery.Result, result *ValidationResult) {
	for name, shortcut := range cfg.Shortcuts {
		if len(shortcut.With) == 0 {
			continue
		}
		target, err := disc.Find(shortcut.Run)
		if err != nil {
			continue // broken reference already caught by validateShortcutRefs
		}
		for _, msg := range checkWithInputs(shortcut.With, target) {
			result.Errors = append(result.Errors,
				fmt.Sprintf("pi.yaml: shortcut %q %s", name, msg))
		}
	}
}

// validateSetupInputs checks that with: keys on setup entries match declared
// inputs on the target automation.
func validateSetupInputs(cfg *config.ProjectConfig, disc *discovery.Result, result *ValidationResult) {
	for i, entry := range cfg.Setup {
		if len(entry.With) == 0 {
			continue
		}
		target, err := disc.Find(entry.Run)
		if err != nil {
			continue // broken reference already caught by validateSetupRefs
		}
		for _, msg := range checkWithInputs(entry.With, target) {
			result.Errors = append(result.Errors,
				fmt.Sprintf("pi.yaml: setup[%d] %s", i, msg))
		}
	}
}

// validateRunStepInputs checks that with: keys on run: steps match declared
// inputs on the target automation.
func validateRunStepInputs(disc *discovery.Result, result *ValidationResult) {
	for _, name := range disc.Names() {
		a := disc.Automations[name]
		automation.WalkSteps(a, func(step automation.Step, loc automation.StepLocation) {
			if step.Type != automation.StepTypeRun || len(step.With) == 0 {
				return
			}
			target, err := disc.Find(step.Value)
			if err != nil {
				return // broken reference already caught by validateRunStepRefs
			}
			for _, msg := range checkWithInputs(step.With, target) {
				result.Errors = append(result.Errors,
					fmt.Sprintf("%s %s", loc.FormatPath(a.Name), msg))
			}
		})
	}
}

// checkWithInputs validates that with: keys match the target automation's
// declared inputs. Returns a list of error messages (empty if all valid).
func checkWithInputs(with map[string]string, target *automation.Automation) []string {
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
