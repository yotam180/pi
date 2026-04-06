package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/vyper-tooling/pi/internal/config"
	"github.com/vyper-tooling/pi/internal/executor"
	"github.com/vyper-tooling/pi/internal/validate"
)

func newValidateCmd() *cobra.Command {
	var showWarnings bool

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate all automation files and config",
		Long: `Statically validate pi.yaml and all automation YAML files in .pi/ without
executing anything. Checks for schema errors, broken references, missing
script files, input mismatches, circular dependencies, and configuration mistakes.

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
  - Circular run: step dependencies (A → B → A)
  - if: conditions are syntactically valid with known predicates
  - Unknown fields in automation YAML files (with "did you mean?" suggestions)
  - Unknown fields in pi.yaml (with "did you mean?" suggestions)

Use --warnings to also check for non-fatal issues:
  - Automations without a description: field
  - Local automations not referenced by any shortcut, setup, or run: step
  - Shortcuts whose names shadow shell builtins or common commands

Exits with code 0 if all checks pass, or code 1 if any errors are found.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := getwd()
			if err != nil {
				return err
			}
			return runValidate(cwd, cmd.OutOrStdout(), cmd.ErrOrStderr(), showWarnings)
		},
	}

	cmd.Flags().BoolVarP(&showWarnings, "warnings", "w", false, "Include non-fatal warnings in output")

	return cmd
}

func runValidate(startDir string, stdout, stderr io.Writer, showWarnings bool) error {
	pc, err := resolveProject(startDir)
	if err != nil {
		return err
	}

	result := validateProject(pc.Root, showWarnings)

	if len(result.Errors) > 0 {
		for _, e := range result.Errors {
			fmt.Fprintf(stderr, "✗ %s\n", e)
		}
		fmt.Fprintf(stderr, "\nValidation failed: %d error(s)\n", len(result.Errors))
		return &executor.ExitError{Code: 1}
	}

	if len(result.Warnings) > 0 {
		for _, w := range result.Warnings {
			fmt.Fprintf(stderr, "⚠ %s\n", w)
		}
		fmt.Fprintln(stderr)
	}

	msg := fmt.Sprintf("✓ Validated %d automation(s), %d shortcut(s), %d setup entry(ies)",
		result.AutomationCount, result.ShortcutCount, result.SetupCount)
	if len(result.Warnings) > 0 {
		msg += fmt.Sprintf(", %d warning(s)", len(result.Warnings))
	}
	fmt.Fprintln(stdout, msg)
	return nil
}

func validateProject(root string, includeWarnings bool) validate.Result {
	cfg, cfgErr := config.Load(root)
	if cfgErr != nil {
		return validate.Result{
			Errors: []string{fmt.Sprintf("pi.yaml: %s", cfgErr)},
		}
	}

	disc, discErr := discoverAllWithConfig(root, cfg, nil)
	if discErr != nil {
		return validate.Result{
			Errors: []string{fmt.Sprintf(".pi/: %s", discErr)},
		}
	}

	ctx := &validate.Context{
		Root:      root,
		Config:    cfg,
		Discovery: disc,
	}

	return validate.DefaultRunner().RunWithOpts(ctx, includeWarnings)
}
