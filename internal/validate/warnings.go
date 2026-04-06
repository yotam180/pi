package validate

import (
	"fmt"
	"sort"

	"github.com/vyper-tooling/pi/internal/automation"
	"github.com/vyper-tooling/pi/internal/shell"
)

// warnMissingDescription flags local automations without a description: field.
// Descriptions power pi list, pi info, and AI-assisted workflows.
func warnMissingDescription(ctx *Context) []string {
	var warns []string
	for _, name := range ctx.Discovery.Names() {
		if ctx.Discovery.IsBuiltin(name) || ctx.Discovery.IsPackage(name) {
			continue
		}
		a := ctx.Discovery.Automations[name]
		if a.Description == "" {
			warns = append(warns, fmt.Sprintf("%s: missing description", name))
		}
	}
	return warns
}

// warnUnusedAutomations flags local automations that are not referenced by
// any shortcut, setup entry, or run: step from any other automation.
func warnUnusedAutomations(ctx *Context) []string {
	referenced := make(map[string]bool)

	for _, sc := range ctx.Config.Shortcuts {
		if target, err := ctx.Discovery.Find(sc.Run); err == nil {
			referenced[target.Name] = true
		}
	}

	for _, entry := range ctx.Config.Setup {
		if target, err := ctx.Discovery.Find(entry.Run); err == nil {
			referenced[target.Name] = true
		}
	}

	for _, name := range ctx.Discovery.Names() {
		a := ctx.Discovery.Automations[name]
		automation.WalkSteps(a, func(step automation.Step, loc automation.StepLocation) {
			if step.Type != automation.StepTypeRun {
				return
			}
			if target, err := ctx.Discovery.Find(step.Value); err == nil {
				referenced[target.Name] = true
			}
		})
	}

	var warns []string
	for _, name := range ctx.Discovery.Names() {
		if ctx.Discovery.IsBuiltin(name) || ctx.Discovery.IsPackage(name) {
			continue
		}
		if !referenced[name] {
			warns = append(warns, fmt.Sprintf("%s: not referenced by any shortcut, setup entry, or run: step", name))
		}
	}
	sort.Strings(warns)
	return warns
}

// warnShortcutShadowing flags shortcuts whose names shadow shell builtins or
// common system commands. Already checked at pi shell install time, but
// checking here catches issues before the developer runs pi shell.
func warnShortcutShadowing(ctx *Context) []string {
	if len(ctx.Config.Shortcuts) == 0 {
		return nil
	}

	names := make([]string, 0, len(ctx.Config.Shortcuts))
	for name := range ctx.Config.Shortcuts {
		names = append(names, name)
	}

	warnings := shell.CheckShadowedNames(names)
	if len(warnings) == 0 {
		return nil
	}

	var warns []string
	for _, w := range warnings {
		warns = append(warns, fmt.Sprintf("pi.yaml: %s", shell.FormatWarning(w)))
	}
	return warns
}
