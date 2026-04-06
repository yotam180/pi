package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vyper-tooling/pi/internal/config"
	"github.com/vyper-tooling/pi/internal/display"
	"github.com/vyper-tooling/pi/internal/executor"
	"github.com/vyper-tooling/pi/internal/project"
	"github.com/vyper-tooling/pi/internal/tools"
)

// setupAddKnownTools returns the short-name-to-builtin map from the tool registry.
func setupAddKnownTools() map[string]string {
	return tools.BuildShortNameMap()
}

func newSetupAddCmd() *cobra.Command {
	var versionFlag string
	var ifFlag string
	var sourceFlag string
	var groupsFlag string
	var yesFlag bool
	var onlyAddFlag bool

	cmd := &cobra.Command{
		Use:   "add <name> [key=value ...]",
		Short: "Add a setup entry to pi.yaml",
		Long: fmt.Sprintf(`Add a setup automation to the setup section of pi.yaml.

By default, the automation is executed first. If it succeeds, the entry is
written to pi.yaml. If it fails, pi.yaml is not modified.

Use --only-add to skip execution and write the entry directly (useful for CI
bootstrapping or when the tool is already set up).

%s
Examples:
  pi setup add python --version 3.13
  pi setup add uv
  pi setup add pi:install-node --version 22
  pi setup add setup/install-deps
  pi setup add pi:install-homebrew --if os.macos
  pi setup add pi:cursor/install-extensions file=.pi/cursor/extensions.txt
  pi setup add uv --only-add`, tools.ToolResolutionHelp()),
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := getwd()
			if err != nil {
				return err
			}
			return runSetupAdd(cwd, args[0], args[1:], versionFlag, ifFlag, sourceFlag, groupsFlag, yesFlag, onlyAddFlag, os.Stdin, os.Stdout, os.Stderr)
		},
	}

	cmd.Flags().StringVar(&versionFlag, "version", "", "sets with: version: \"<v>\"")
	cmd.Flags().StringVar(&ifFlag, "if", "", "sets if: <expr> on the setup entry")
	cmd.Flags().StringVar(&sourceFlag, "source", "", "sets with: source: \"<path>\"")
	cmd.Flags().StringVar(&groupsFlag, "groups", "", "sets with: groups: \"<list>\"")
	cmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "skip all prompts (auto-confirm)")
	cmd.Flags().BoolVar(&onlyAddFlag, "only-add", false, "skip execution and write the entry directly")

	return cmd
}

func runSetupAdd(root, name string, kvArgs []string, versionFlag, ifFlag, sourceFlag, groupsFlag string, yesFlag, onlyAdd bool, stdin io.Reader, stdout, stderr io.Writer) error {
	printer := display.NewForWriter(stdout)

	piYamlExists := fileExists(root, config.FileName)
	if !piYamlExists {
		ok, err := promptInit(root, yesFlag, stdin, stdout)
		if err != nil {
			return err
		}
		if !ok {
			fmt.Fprintln(stderr, "Aborted.")
			return &executor.ExitError{Code: 1}
		}
	}

	resolvedName := name
	if expanded, ok := setupAddKnownTools()[name]; ok {
		resolvedName = expanded
		printer.Dim("Resolved '%s' → %s\n", name, resolvedName)
		fmt.Fprintln(stdout)
	}

	entry := config.SetupEntry{
		Run: resolvedName,
		If:  ifFlag,
	}

	withMap := make(map[string]string)
	if versionFlag != "" {
		withMap["version"] = versionFlag
	}
	if sourceFlag != "" {
		withMap["source"] = sourceFlag
	}
	if groupsFlag != "" {
		withMap["groups"] = groupsFlag
	}

	for _, kv := range kvArgs {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 || parts[0] == "" {
			return fmt.Errorf("invalid key=value argument: %q (expected key=value)", kv)
		}
		withMap[parts[0]] = parts[1]
	}

	if len(withMap) > 0 {
		entry.With = withMap
	}

	if !onlyAdd {
		if err := invokeSetupAutomation(root, entry, stdout, stderr); err != nil {
			return err
		}
	}

	if err := config.AddSetupEntry(root, entry); err != nil {
		var dupErr *config.DuplicateSetupEntryError
		if errors.As(err, &dupErr) {
			fmt.Fprintln(stdout, "Already in pi.yaml. No changes made.")
			return nil
		}
		var replErr *config.ReplacedSetupEntryError
		if errors.As(err, &replErr) {
			fmt.Fprintln(stdout, "Replaced in pi.yaml:")
			fmt.Fprintln(stdout, config.FormatSetupEntry(entry))
			return nil
		}
		return err
	}

	fmt.Fprintln(stdout, "Added to setup in pi.yaml:")
	fmt.Fprintln(stdout, config.FormatSetupEntry(entry))

	return nil
}

// invokeSetupAutomation runs a single setup automation before writing it to
// pi.yaml. It reuses the same resolution and execution pipeline as pi setup.
// Stdout/stderr are streamed live to the caller's terminal.
func invokeSetupAutomation(root string, entry config.SetupEntry, stdout, stderr io.Writer) error {
	projRoot, err := project.FindRoot(root)
	if err != nil {
		projRoot = root
	}

	cfg, _ := config.Load(projRoot)
	pc := &ProjectContext{Root: projRoot, Config: cfg}

	result, err := pc.Discover(stderr)
	if err != nil {
		return fmt.Errorf("discovery failed: %w", err)
	}

	if entry.If != "" {
		skip, err := evaluateSetupCondition(entry.If, projRoot)
		if err != nil {
			return fmt.Errorf("if: %w", err)
		}
		if skip {
			return nil
		}
	}

	a, err := result.Find(entry.Run)
	if err != nil {
		return fmt.Errorf("automation %q: %w", entry.Run, err)
	}

	exec := pc.NewExecutor(result, ExecutorOpts{
		Stdout: stdout,
		Stderr: stderr,
	})

	if err := exec.RunWithInputs(a, nil, entry.With); err != nil {
		return fmt.Errorf("automation %q failed: %w", entry.Run, err)
	}
	fmt.Fprintln(stdout)

	return nil
}

// promptInit asks the user to initialize a PI project when pi.yaml doesn't exist.
func promptInit(root string, yesFlag bool, stdin io.Reader, stdout io.Writer) (bool, error) {
	inferred := toKebabCase(inferDirName(root))

	fmt.Fprintln(stdout, "No pi.yaml found.")
	fmt.Fprintln(stdout)

	if yesFlag || !isInteractive(stdin) {
		if err := initProject(root, inferred, stdout); err != nil {
			return false, err
		}
		fmt.Fprintln(stdout)
		return true, nil
	}

	fmt.Fprintf(stdout, "Initialize project '%s'? [Y/n] ", inferred)

	reader := bufio.NewReader(stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, nil
	}

	input = strings.TrimSpace(strings.ToLower(input))
	if input != "" && input != "y" && input != "yes" {
		return false, nil
	}

	if err := initProject(root, inferred, stdout); err != nil {
		return false, err
	}
	fmt.Fprintln(stdout)
	return true, nil
}

func inferDirName(root string) string {
	base := strings.TrimRight(root, "/\\")
	idx := strings.LastIndexAny(base, "/\\")
	if idx >= 0 {
		return base[idx+1:]
	}
	return base
}

func fileExists(dir, name string) bool {
	_, err := os.Stat(filepath.Join(dir, name))
	return err == nil
}
