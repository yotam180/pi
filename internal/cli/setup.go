package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/vyper-tooling/pi/internal/automation"
	"github.com/vyper-tooling/pi/internal/conditions"
	"github.com/vyper-tooling/pi/internal/config"
	"github.com/vyper-tooling/pi/internal/display"
	"github.com/vyper-tooling/pi/internal/executor"
	"github.com/vyper-tooling/pi/internal/project"
	"github.com/vyper-tooling/pi/internal/runtimes"
	"github.com/vyper-tooling/pi/internal/shell"
)

var ciEnvVars = []string{
	"CI",
	"GITHUB_ACTIONS",
	"GITLAB_CI",
	"CIRCLECI",
	"JENKINS_URL",
	"BUILDKITE",
	"TRAVIS",
	"CODEBUILD_BUILD_ID",
	"TF_BUILD",
}

func newSetupCmd() *cobra.Command {
	var noShell bool
	var silent bool
	var loud bool

	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Run all setup automations",
		Long: `Run all automations listed in the setup section of pi.yaml sequentially.
After setup completes, pi shell is run automatically to install shortcuts
(skipped in CI environments or with --no-shell).
Use --silent to suppress PI status lines for installer automations.
Use --loud to force trace lines and output for all steps (overrides silent: true).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetup(cmd.OutOrStdout(), cmd.ErrOrStderr(), noShell, silent, loud)
		},
	}

	cmd.Flags().BoolVar(&noShell, "no-shell", false, "skip shell shortcut installation")
	cmd.Flags().BoolVar(&silent, "silent", false, "suppress PI status lines for installer automations")
	cmd.Flags().BoolVar(&loud, "loud", false, "force trace lines and output for all steps (overrides silent: true)")

	return cmd
}

func runSetup(stdout, stderr io.Writer, noShell, silent, loud bool) error {
	automation.WarnWriter = stderr
	defer func() { automation.WarnWriter = nil }()

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	root, err := project.FindRoot(cwd)
	if err != nil {
		return err
	}

	cfg, err := config.Load(root)
	if err != nil {
		return err
	}

	out := display.New(stdout)
	stderrPrinter := display.New(stderr)

	if len(cfg.Setup) == 0 && len(cfg.Shortcuts) == 0 {
		fmt.Fprintln(stdout, "Nothing to set up (no setup entries or shortcuts in pi.yaml).")
		return nil
	}

	if len(cfg.Setup) > 0 {
		result, err := discoverAll(root)
		if err != nil {
			return err
		}

		exec := &executor.Executor{
			RepoRoot:       root,
			Discovery:      result,
			Stdout:         stdout,
			Stderr:         stderr,
			Silent:         silent,
			Loud:           loud,
			Printer:        stderrPrinter,
			ParentEvalFile: os.Getenv("PI_PARENT_EVAL_FILE"),
		}

		if cfg.EffectiveProvisionMode() != config.ProvisionNever {
			exec.Provisioner = runtimes.NewProvisioner(cfg, stderr)
		}

		for i, entry := range cfg.Setup {
			if entry.If != "" {
				skip, err := evaluateSetupCondition(entry.If, root)
				if err != nil {
					return fmt.Errorf("setup[%d] if: %w", i, err)
				}
				if skip {
					out.SetupHeader("==> setup[%d]: %s [skipped] (condition: %s)\n", i, entry.Run, entry.If)
					continue
				}
			}
			out.SetupHeader("==> setup[%d]: %s\n", i, entry.Run)
			a, err := result.Find(entry.Run)
			if err != nil {
				return fmt.Errorf("setup[%d]: %w", i, err)
			}
			if err := exec.RunWithInputs(a, nil, entry.With); err != nil {
				return fmt.Errorf("setup[%d] %q failed: %w", i, entry.Run, err)
			}
		}
	}

	skipShell := noShell || isCI()
	if skipShell {
		if isCI() {
			fmt.Fprintln(stdout, "Skipping shell shortcuts (CI environment detected).")
		} else {
			fmt.Fprintln(stdout, "Skipping shell shortcuts (--no-shell).")
		}
		return nil
	}

	if len(cfg.Shortcuts) > 0 {
		out.SetupHeader("==> Installing shell shortcuts...\n")
		piBin, err := resolvePiBinary()
		if err != nil {
			return err
		}
		shellPath, err := shell.Install(cfg, piBin, root)
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "Installed %d shortcut(s) to %s\n", len(cfg.Shortcuts), shellPath)

		if evalFile := os.Getenv("PI_PARENT_EVAL_FILE"); evalFile != "" {
			if rc := shell.PrimaryRCFile(); rc != "" {
				if err := executor.AppendToParentEval(evalFile, fmt.Sprintf("source %q", rc)); err != nil {
					fmt.Fprintf(stderr, "warning: could not write parent eval file: %v\n", err)
				}
			}
		}
	}

	out.Green("Setup complete.\n")
	return nil
}

// evaluateSetupCondition resolves and evaluates an if: expression for a setup entry.
// Returns true if the entry should be skipped (condition evaluates to false).
func evaluateSetupCondition(expr string, repoRoot string) (bool, error) {
	predNames, err := conditions.Predicates(expr)
	if err != nil {
		return false, err
	}

	resolved, err := executor.ResolvePredicates(predNames, repoRoot)
	if err != nil {
		return false, err
	}

	result, err := conditions.Eval(expr, resolved)
	if err != nil {
		return false, err
	}

	return !result, nil
}

func isCI() bool {
	for _, v := range ciEnvVars {
		if os.Getenv(v) != "" {
			return true
		}
	}
	return false
}
