package validate

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Known top-level YAML keys in pi.yaml.
var knownPiYamlKeys = map[string]bool{
	"project":   true,
	"shortcuts": true,
	"setup":     true,
	"packages":  true,
	"runtimes":  true,
}

// Known keys inside the runtimes: block.
var knownRuntimesKeys = map[string]bool{
	"provision": true,
	"manager":   true,
}

func checkPiYamlUnknownFields(ctx *Context) []string {
	piYamlPath := filepath.Join(ctx.Root, "pi.yaml")
	data, err := os.ReadFile(piYamlPath)
	if err != nil {
		return nil
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil
	}

	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		return nil
	}
	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil
	}

	var errs []string

	for i := 0; i < len(root.Content)-1; i += 2 {
		keyNode := root.Content[i]
		valueNode := root.Content[i+1]
		key := keyNode.Value

		if !knownPiYamlKeys[key] {
			msg := fmt.Sprintf("pi.yaml: unknown field %q", key)
			if suggestion := suggestField(key, knownPiYamlKeys); suggestion != "" {
				msg += fmt.Sprintf(" (did you mean %q?)", suggestion)
			}
			errs = append(errs, msg)
			continue
		}

		if key == "runtimes" && valueNode.Kind == yaml.MappingNode {
			runtimeErrs := checkRuntimesNodeUnknownFields(valueNode)
			errs = append(errs, runtimeErrs...)
		}
	}

	return errs
}

func checkRuntimesNodeUnknownFields(node *yaml.Node) []string {
	var errs []string
	for i := 0; i < len(node.Content)-1; i += 2 {
		keyNode := node.Content[i]
		key := keyNode.Value

		if !knownRuntimesKeys[key] {
			msg := fmt.Sprintf("pi.yaml runtimes: unknown field %q", key)
			if suggestion := suggestField(key, knownRuntimesKeys); suggestion != "" {
				msg += fmt.Sprintf(" (did you mean %q?)", suggestion)
			}
			errs = append(errs, msg)
		}
	}
	return errs
}
