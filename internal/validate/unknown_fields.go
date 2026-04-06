package validate

import (
	"fmt"
	"os"
	"sort"

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
			msg := formatUnknownFieldError(automationName, key, sortedKeys(knownAutomationKeys))
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
			msg := formatUnknownStepFieldError(automationName, stepIndex, key, sortedKeys(knownStepKeys))
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
			if suggestion := suggestField(key, knownStepKeys); suggestion != "" {
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
			if suggestion := suggestField(key, knownInstallKeys); suggestion != "" {
				msg += fmt.Sprintf(" (did you mean %q?)", suggestion)
			}
			errs = append(errs, msg)
		}
	}
	return errs
}

func formatUnknownFieldError(automationName, key string, knownKeys []string) string {
	msg := fmt.Sprintf("%s: unknown field %q", automationName, key)
	known := make(map[string]bool, len(knownKeys))
	for _, k := range knownKeys {
		known[k] = true
	}
	if suggestion := suggestField(key, known); suggestion != "" {
		msg += fmt.Sprintf(" (did you mean %q?)", suggestion)
	}
	return msg
}

func formatUnknownStepFieldError(automationName string, stepIndex int, key string, knownKeys []string) string {
	msg := fmt.Sprintf("%s step[%d]: unknown field %q", automationName, stepIndex, key)
	known := make(map[string]bool, len(knownKeys))
	for _, k := range knownKeys {
		known[k] = true
	}
	if suggestion := suggestField(key, known); suggestion != "" {
		msg += fmt.Sprintf(" (did you mean %q?)", suggestion)
	}
	return msg
}

// suggestField returns the closest known field name to the given unknown key,
// or "" if nothing is close enough. Uses Levenshtein distance.
func suggestField(unknown string, known map[string]bool) string {
	maxDist := len(unknown) / 2
	if maxDist < 2 {
		maxDist = 2
	}

	type match struct {
		name     string
		distance int
	}
	var matches []match

	for name := range known {
		d := levenshtein(unknown, name)
		if d > 0 && d <= maxDist {
			matches = append(matches, match{name: name, distance: d})
		}
	}

	if len(matches) == 0 {
		return ""
	}

	sort.Slice(matches, func(i, j int) bool {
		if matches[i].distance != matches[j].distance {
			return matches[i].distance < matches[j].distance
		}
		return matches[i].name < matches[j].name
	})

	return matches[0].name
}

// levenshtein computes the edit distance between two strings.
func levenshtein(a, b string) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	prev := make([]int, lb+1)
	for j := range prev {
		prev[j] = j
	}

	curr := make([]int, lb+1)
	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			ins := curr[j-1] + 1
			del := prev[j] + 1
			sub := prev[j-1] + cost
			m := ins
			if del < m {
				m = del
			}
			if sub < m {
				m = sub
			}
			curr[j] = m
		}
		prev, curr = curr, prev
	}
	return prev[lb]
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
