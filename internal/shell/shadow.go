package shell

import (
	"fmt"
	"sort"
	"strings"
)

// ShadowWarning describes a shortcut name that shadows a known shell builtin
// or common system command.
type ShadowWarning struct {
	Name       string
	Kind       string // "shell builtin" or "common command"
	Suggestion string
}

// shellBuiltins are POSIX and common shell built-in commands that should never
// be shadowed. Sourced from the POSIX spec and bash/zsh documentation.
var shellBuiltins = map[string]bool{
	"alias":    true,
	"bg":       true,
	"cd":       true,
	"command":  true,
	"echo":     true,
	"eval":     true,
	"exec":     true,
	"exit":     true,
	"export":   true,
	"false":    true,
	"fg":       true,
	"getopts":  true,
	"hash":     true,
	"jobs":     true,
	"kill":     true,
	"local":    true,
	"printf":   true,
	"pwd":      true,
	"read":     true,
	"readonly": true,
	"return":   true,
	"set":      true,
	"shift":    true,
	"source":   true,
	"test":     true,
	"time":     true,
	"times":    true,
	"trap":     true,
	"true":     true,
	"type":     true,
	"ulimit":   true,
	"umask":    true,
	"unalias":  true,
	"unset":    true,
	"wait":     true,
	"which":    true,
}

// commonCommands are widely used system commands that shadowing would likely
// break scripts and other tools.
var commonCommands = map[string]bool{
	"cat":    true,
	"cp":     true,
	"curl":   true,
	"chmod":  true,
	"chown":  true,
	"diff":   true,
	"find":   true,
	"git":    true,
	"grep":   true,
	"head":   true,
	"less":   true,
	"ln":     true,
	"ls":     true,
	"make":   true,
	"man":    true,
	"mkdir":  true,
	"mv":     true,
	"rm":     true,
	"rmdir":  true,
	"run":    true,
	"sed":    true,
	"sort":   true,
	"ssh":    true,
	"sudo":   true,
	"tail":   true,
	"tar":    true,
	"touch":  true,
	"wc":     true,
	"wget":   true,
	"xargs":  true,
}

// CheckShadowedNames checks a list of shortcut names against known shell
// builtins and common system commands. Returns a warning for each name that
// shadows a known command, sorted by shortcut name.
func CheckShadowedNames(names []string) []ShadowWarning {
	var warnings []ShadowWarning
	for _, name := range names {
		lower := strings.ToLower(name)
		if shellBuiltins[lower] {
			warnings = append(warnings, ShadowWarning{
				Name:       name,
				Kind:       "shell builtin",
				Suggestion: suggestAlternative(name),
			})
		} else if commonCommands[lower] {
			warnings = append(warnings, ShadowWarning{
				Name:       name,
				Kind:       "common command",
				Suggestion: suggestAlternative(name),
			})
		}
	}
	sort.Slice(warnings, func(i, j int) bool {
		return warnings[i].Name < warnings[j].Name
	})
	return warnings
}

// FormatWarning returns a human-readable warning string for a single shadow warning.
func FormatWarning(w ShadowWarning) string {
	msg := fmt.Sprintf("warning: shortcut %q shadows the %s %q", w.Name, w.Kind, w.Name)
	if w.Suggestion != "" {
		msg += fmt.Sprintf(" — consider renaming to %q", w.Suggestion)
	}
	return msg
}

func suggestAlternative(name string) string {
	if len(name) <= 2 {
		return "pi-" + name
	}
	return "pi-" + name
}
