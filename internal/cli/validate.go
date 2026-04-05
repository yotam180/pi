package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/vyper-tooling/pi/internal/automation"
	"github.com/vyper-tooling/pi/internal/config"
	"github.com/vyper-tooling/pi/internal/discovery"
	"github.com/vyper-tooling/pi/internal/executor"
	"github.com/vyper-tooling/pi/internal/project"
)

func newValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate all automation files and config",
		Long: `Statically validate pi.yaml and all automation YAML files in .pi/ without
executing anything. Checks for schema errors, broken references, and
configuration mistakes.

Checks performed:
  - pi.yaml parsing and schema validation
  - All .pi/*.yaml automation parsing and schema validation
  - Shortcut targets reference existing automations
  - Setup entry targets reference existing automations
  - run: steps reference existing automations

Exits with code 0 if all checks pass, or code 1 if any errors are found.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
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
	root, err := project.FindRoot(startDir)
	if err != nil {
		return err
	}

	automation.WarnWriter = stderr
	defer func() { automation.WarnWriter = nil }()

	result := validateProject(root)

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
		validateAutomationStepRefs(a, disc, result)
	}
}

func validateAutomationStepRefs(a *automation.Automation, disc *discovery.Result, result *ValidationResult) {
	for i, step := range a.Steps {
		if step.IsFirst() {
			for j, sub := range step.First {
				if sub.Type != automation.StepTypeRun {
					continue
				}
				if _, err := disc.Find(sub.Value); err != nil {
					result.Errors = append(result.Errors,
						fmt.Sprintf("%s: step[%d].first[%d] run: references unknown automation %q", a.Name, i, j, sub.Value))
				}
			}
			continue
		}
		if step.Type != automation.StepTypeRun {
			continue
		}
		if _, err := disc.Find(step.Value); err != nil {
			result.Errors = append(result.Errors,
				fmt.Sprintf("%s: step[%d] run: references unknown automation %q", a.Name, i, step.Value))
		}
	}

	if a.Install == nil {
		return
	}
	validatePhaseStepRefs(a.Name, "test", &a.Install.Test, disc, result)
	validatePhaseStepRefs(a.Name, "run", &a.Install.Run, disc, result)
	if a.Install.Verify != nil {
		validatePhaseStepRefs(a.Name, "verify", a.Install.Verify, disc, result)
	}
}

func validatePhaseStepRefs(automationName, phaseName string, phase *automation.InstallPhase, disc *discovery.Result, result *ValidationResult) {
	if phase.IsScalar {
		return
	}
	for i, step := range phase.Steps {
		if step.IsFirst() {
			for j, sub := range step.First {
				if sub.Type != automation.StepTypeRun {
					continue
				}
				if _, err := disc.Find(sub.Value); err != nil {
					result.Errors = append(result.Errors,
						fmt.Sprintf("%s: install.%s step[%d].first[%d] run: references unknown automation %q",
							automationName, phaseName, i, j, sub.Value))
				}
			}
			continue
		}
		if step.Type != automation.StepTypeRun {
			continue
		}
		if _, err := disc.Find(step.Value); err != nil {
			result.Errors = append(result.Errors,
				fmt.Sprintf("%s: install.%s step[%d] run: references unknown automation %q",
					automationName, phaseName, i, step.Value))
		}
	}
}
