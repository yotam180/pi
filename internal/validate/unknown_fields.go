package validate

import (
	"fmt"
	"maps"
	"os"
	"slices"

	"github.com/vyper-tooling/pi/internal/suggest"
	"gopkg.in/yaml.v3"
)

// Known YAML keys at the automation (top) level.
var knownAutomationKeys = map[string]bool{
	"name":         true,
	"description":  true,
	"steps":        true,
	"install":      true,
	"inputs":       true,
	"if":           true,
	"requires":     true,
	"bash":         true,
	"python":       true,
	"typescript":   true,
	"run":          true,
	"env":          true,
	"dir":          true,
	"timeout":      true,
	"silent":       true,
	"parent_shell": true,
	"pipe_to":      true,
	"pipe":         true,
	"with":         true,
}

// Known YAML keys at the step level (inside steps: list entries).
var knownStepKeys = map[string]bool{
	"bash":         true,
	"python":       true,
	"typescript":   true,
	"run":          true,
	"if":           true,
	"env":          true,
	"dir":          true,
	"timeout":      true,
	"silent":       true,
	"parent_shell": true,
	"pipe_to":      true,
	"pipe":         true,
	"description":  true,
	"first":        true,
	"with":         true,
}

// Known YAML keys inside install: blocks.
var knownInstallKeys = map[string]bool{
	"test":    true,
	"run":     true,
	"verify":  true,
	"version": true,
}

func checkUnknownFields(ctx *Context) []string {
	var errs []string
	for _, name := range ctx.Discovery.Names() {
		if ctx.Discovery.IsBuiltin(name) {
			continue
		}
		if ctx.Discovery.IsPackage(name) {
			continue
		}

		a := ctx.Discovery.Automations[name]
		if a.FilePath == "" {
			continue
		}

		fileErrs := checkFileUnknownFields(a.FilePath, name)
		errs = append(errs, fileErrs...)
	}
	return errs
}

func checkFileUnknownFields(filePath, automationName string) []string {
	data, err := os.ReadFile(filePath)
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

	// Check top-level keys
	for i := 0; i < len(root.Content)-1; i += 2 {
		keyNode := root.Content[i]
		valueNode := root.Content[i+1]
		key := keyNode.Value

		if !knownAutomationKeys[key] {
			msg := formatUnknownFieldError(automationName, key, slices.Sorted(maps.Keys(knownAutomationKeys)))
			errs = append(errs, msg)
			continue
		}

		if key == "steps" && valueNode.Kind == yaml.SequenceNode {
			for j, stepNode := range valueNode.Content {
				stepErrs := checkStepNodeUnknownFields(automationName, stepNode, j)
				errs = append(errs, stepErrs...)
			}
		}

		if key == "install" && valueNode.Kind == yaml.MappingNode {
			installErrs := checkInstallNodeUnknownFields(automationName, valueNode)
			errs = append(errs, installErrs...)
		}
	}

	return errs
}

func checkStepNodeUnknownFields(automationName string, node *yaml.Node, stepIndex int) []string {
	if node.Kind != yaml.MappingNode {
		return nil
	}

	var errs []string
	for i := 0; i < len(node.Content)-1; i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]
		key := keyNode.Value

		if !knownStepKeys[key] {
			msg := formatUnknownStepFieldError(automationName, stepIndex, key, slices.Sorted(maps.Keys(knownStepKeys)))
			errs = append(errs, msg)
			continue
		}

		if key == "first" && valueNode.Kind == yaml.SequenceNode {
			for j, subNode := range valueNode.Content {
				subErrs := checkFirstSubStepUnknownFields(automationName, stepIndex, subNode, j)
				errs = append(errs, subErrs...)
			}
		}
	}
	return errs
}

func checkFirstSubStepUnknownFields(automationName string, stepIndex int, node *yaml.Node, subIndex int) []string {
	if node.Kind != yaml.MappingNode {
		return nil
	}

	var errs []string
	for i := 0; i < len(node.Content)-1; i += 2 {
		keyNode := node.Content[i]
		key := keyNode.Value

		if !knownStepKeys[key] {
			msg := fmt.Sprintf("%s step[%d].first[%d]: unknown field %q", automationName, stepIndex, subIndex, key)
			if suggestion := suggestFieldName(key, knownStepKeys); suggestion != "" {
				msg += fmt.Sprintf(" (did you mean %q?)", suggestion)
			}
			errs = append(errs, msg)
		}
	}
	return errs
}

func checkInstallNodeUnknownFields(automationName string, node *yaml.Node) []string {
	if node.Kind != yaml.MappingNode {
		return nil
	}

	var errs []string
	for i := 0; i < len(node.Content)-1; i += 2 {
		keyNode := node.Content[i]
		key := keyNode.Value

		if !knownInstallKeys[key] {
			msg := fmt.Sprintf("%s install: unknown field %q", automationName, key)
			if suggestion := suggestFieldName(key, knownInstallKeys); suggestion != "" {
				msg += fmt.Sprintf(" (did you mean %q?)", suggestion)
			}
			errs = append(errs, msg)
		}
	}
	return errs
}

func formatUnknownFieldError(automationName, key string, knownKeys []string) string {
	msg := fmt.Sprintf("%s: unknown field %q", automationName, key)
	if s := suggest.Best(key, knownKeys, fieldSuggestMaxDist(key)); s != "" {
		msg += fmt.Sprintf(" (did you mean %q?)", s)
	}
	return msg
}

func formatUnknownStepFieldError(automationName string, stepIndex int, key string, knownKeys []string) string {
	msg := fmt.Sprintf("%s step[%d]: unknown field %q", automationName, stepIndex, key)
	if s := suggest.Best(key, knownKeys, fieldSuggestMaxDist(key)); s != "" {
		msg += fmt.Sprintf(" (did you mean %q?)", s)
	}
	return msg
}

// suggestFieldName returns the closest known field name to the given unknown
// key, or "" if nothing is close enough.
func suggestFieldName(unknown string, known map[string]bool) string {
	candidates := slices.Sorted(maps.Keys(known))
	return suggest.Best(unknown, candidates, fieldSuggestMaxDist(unknown))
}

func fieldSuggestMaxDist(field string) int {
	maxDist := len(field) / 2
	if maxDist < 2 {
		maxDist = 2
	}
	return maxDist
}
