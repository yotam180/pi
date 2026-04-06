package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vyper-tooling/pi/internal/display"
)

func newNewCmd() *cobra.Command {
	var bashFlag string
	var pythonFlag string
	var description string

	cmd := &cobra.Command{
		Use:   "new <name>",
		Short: "Scaffold a new automation file",
		Long: `Create a new automation YAML file in the .pi/ directory.

The name determines the file path: "build" creates .pi/build.yaml,
"setup/install-deps" creates .pi/setup/install-deps.yaml.

Examples:
  pi new build                          create .pi/build.yaml
  pi new setup/install-deps             create .pi/setup/install-deps.yaml
  pi new build --bash "cargo build"     pre-fill with a bash command
  pi new fmt --python "format.py"       pre-fill with a python script
  pi new test --description "Run tests" set the description`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := getwd()
			if err != nil {
				return err
			}
			return runNew(cwd, args[0], bashFlag, pythonFlag, description, os.Stdout, os.Stderr)
		},
	}

	cmd.Flags().StringVar(&bashFlag, "bash", "", "pre-fill with a bash command")
	cmd.Flags().StringVar(&pythonFlag, "python", "", "pre-fill with a python script path")
	cmd.Flags().StringVarP(&description, "description", "d", "", "automation description")

	return cmd
}

// runNew implements the pi new logic. Extracted for testability.
func runNew(startDir, name, bash, python, desc string, stdout, stderr io.Writer) error {
	if name == "" {
		return fmt.Errorf("automation name is required")
	}

	name = strings.TrimSuffix(name, ".yaml")
	name = strings.TrimSuffix(name, ".yml")

	root, piDir, err := findPiDir(startDir)
	if err != nil {
		return err
	}

	targetPath := filepath.Join(piDir, name+".yaml")

	if _, err := os.Stat(targetPath); err == nil {
		return fmt.Errorf("automation already exists: .pi/%s.yaml", name)
	}

	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return fmt.Errorf("creating directories: %w", err)
	}

	content := generateAutomationYAML(name, bash, python, desc)

	if err := os.WriteFile(targetPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	relPath, _ := filepath.Rel(root, targetPath)
	if relPath == "" {
		relPath = targetPath
	}

	printer := display.NewForWriter(stdout)
	printer.Green("Created %s\n", relPath)
	fmt.Fprintln(stdout)
	printer.Dim("Next steps:\n")
	printer.Dim("  pi run %s                  run the automation\n", name)
	printer.Dim("  pi info %s                 view automation details\n", name)

	return nil
}

// findPiDir locates the project root (via pi.yaml) and returns
// the root path and .pi/ directory path. Returns an error with
// guidance if no project is found.
func findPiDir(startDir string) (root string, piDir string, err error) {
	pc, err := resolveProject(startDir)
	if err != nil {
		return "", "", fmt.Errorf("%w\n\nRun 'pi init' to initialize a project first", err)
	}

	piDir = filepath.Join(pc.Root, ".pi")
	if err := os.MkdirAll(piDir, 0o755); err != nil {
		return "", "", fmt.Errorf("creating .pi/: %w", err)
	}

	return pc.Root, piDir, nil
}

// generateAutomationYAML builds the YAML content for a new automation file.
func generateAutomationYAML(name, bash, python, desc string) string {
	if desc == "" {
		desc = "TODO — describe what this automation does"
	}

	var b strings.Builder
	if yamlNeedsQuoting(desc) {
		b.WriteString("description: \"" + strings.ReplaceAll(desc, "\"", "\\\"") + "\"\n")
	} else {
		b.WriteString("description: " + desc + "\n")
	}

	switch {
	case bash != "":
		b.WriteString("bash: " + bash + "\n")
	case python != "":
		b.WriteString("python: " + python + "\n")
	default:
		b.WriteString("bash: echo \"hello from " + filepath.Base(name) + "\"\n")
	}

	return b.String()
}

// yamlNeedsQuoting returns true if a string value contains characters
// that require quoting in a YAML scalar context.
func yamlNeedsQuoting(s string) bool {
	return strings.ContainsAny(s, ":#{}[]|>&*!%@`")
}

// ExampleAutomationContent is the YAML content for the example automation
// that pi init creates. Shared so tests can verify the content.
const ExampleAutomationContent = `description: A sample automation — edit or delete this file
bash: echo "Hello from PI!"
`
