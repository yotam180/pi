package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vyper-tooling/pi/internal/config"
	"github.com/vyper-tooling/pi/internal/display"
)

func newInitCmd() *cobra.Command {
	var nameFlag string
	var yesFlag bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new PI project",
		Long: `Initialize a new PI project by creating pi.yaml and the .pi/ directory.

By default, the project name is inferred from the current directory name.
The developer can accept or override via an interactive prompt.

Examples:
  pi init
  pi init --name my-project
  pi init --yes`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := getwd()
			if err != nil {
				return err
			}
			return runInit(cwd, nameFlag, yesFlag, os.Stdin, os.Stdout, os.Stderr)
		},
	}

	cmd.Flags().StringVar(&nameFlag, "name", "", "project name (skips interactive prompt)")
	cmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "accept inferred project name without prompting")

	return cmd
}

// initProject creates pi.yaml and .pi/ in root with the given project name.
// Exported logic so other commands (e.g. setup add) can reuse it.
func initProject(root, name string, stdout io.Writer) error {
	piYamlPath := filepath.Join(root, config.FileName)
	piDirPath := filepath.Join(root, ".pi")

	if err := os.WriteFile(piYamlPath, []byte("project: "+name+"\n"), 0o644); err != nil {
		return fmt.Errorf("creating %s: %w", config.FileName, err)
	}

	if err := os.MkdirAll(piDirPath, 0o755); err != nil {
		return fmt.Errorf("creating .pi/: %w", err)
	}

	printer := display.NewForWriter(stdout)
	printer.Green("Initialized project '%s'.\n", name)
	fmt.Fprintln(stdout)
	printer.Dim("  Created %s\n", config.FileName)
	printer.Dim("  Created .pi/\n")

	printNextSteps(stdout)
	return nil
}

// runInit implements the pi init logic.
func runInit(root, nameFlag string, yesFlag bool, stdin io.Reader, stdout, stderr io.Writer) error {
	piYamlPath := filepath.Join(root, config.FileName)

	if _, err := os.Stat(piYamlPath); err == nil {
		cfg, loadErr := config.Load(root)
		projectName := "<unknown>"
		if loadErr == nil && cfg != nil {
			projectName = cfg.Project
		}

		printer := display.NewForWriter(stdout)
		printer.Dim("Already initialized (project: %s).\n", projectName)
		printNextSteps(stdout)
		return nil
	}

	name, err := resolveProjectName(root, nameFlag, yesFlag, stdin, stdout)
	if err != nil {
		return err
	}

	return initProject(root, name, stdout)
}

// resolveProjectName determines the project name from flags, prompt, or inference.
func resolveProjectName(root, nameFlag string, yesFlag bool, stdin io.Reader, stdout io.Writer) (string, error) {
	if nameFlag != "" {
		return nameFlag, nil
	}

	inferred := toKebabCase(filepath.Base(root))

	if yesFlag || !isInteractive(stdin) {
		return inferred, nil
	}

	return promptProjectName(inferred, stdin, stdout)
}

func promptProjectName(inferred string, stdin io.Reader, stdout io.Writer) (string, error) {
	fmt.Fprintln(stdout, "Initializing PI project.")
	fmt.Fprintln(stdout)
	fmt.Fprintf(stdout, "Project name [%s]: ", inferred)

	reader := bufio.NewReader(stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return inferred, nil
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return inferred, nil
	}

	return input, nil
}

func printNextSteps(w io.Writer) {
	printer := display.NewForWriter(w)
	fmt.Fprintln(w)
	printer.Dim("Next steps:\n")
	printer.Dim("  pi setup add python --version 3.13   add a setup step\n")
	printer.Dim("  pi shell                             install shell shortcuts\n")
	printer.Dim("  pi run <name>                        run an automation\n")
}

var kebabRe = regexp.MustCompile(`[^a-z0-9-]+`)

// toKebabCase converts a directory name to kebab-case.
func toKebabCase(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "_", "-")
	name = kebabRe.ReplaceAllString(name, "")
	name = strings.Trim(name, "-")
	return name
}

// isInteractive returns true when stdin appears to be a terminal.
func isInteractive(r io.Reader) bool {
	f, ok := r.(*os.File)
	if !ok {
		return false
	}
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}
