# `pi setup add` Command

## Type
feature

## Status
todo

## Priority
high

## Project
standalone

## Description
Add a `pi setup add` command that appends an entry to the `setup:` block in `pi.yaml`. The developer types what they want in a natural, short form — `pi setup add python --version 3.13` — and PI resolves their intent, writes the correct YAML, and tells them what it did. If `pi.yaml` doesn't exist, PI offers to initialize the project first.

Read `docs/philosophy.md` before implementing. The entire design philosophy of this command is intent-first: the developer should be able to guess the right command and get it right most of the time.

## Acceptance Criteria
- [ ] `pi setup add pi:install-python --version 3.13` appends `- run: pi:install-python\n  with:\n    version: "3.13"` to `setup:` in `pi.yaml`
- [ ] Short-form tool names are accepted and expanded: `python` → `pi:install-python`, `node` → `pi:install-node`, `go` → `pi:install-go`, `rust` → `pi:install-rust`, `ruby` → `pi:install-ruby`, `uv` → `pi:install-uv`, `tsx` → `pi:install-tsx`, `homebrew` → `pi:install-homebrew`, `terraform` → `pi:install-terraform`, `kubectl` → `pi:install-kubectl`, `helm` → `pi:install-helm`, `pnpm` → `pi:install-pnpm`, `bun` → `pi:install-bun`, `deno` → `pi:install-deno`, `aws-cli` → `pi:install-aws-cli`
- [ ] `pi:<tool>` prefix works too: `pi:python` → `pi:install-python`
- [ ] `--version <v>` flag sets `with: version: "<v>"`
- [ ] `--if <expr>` flag adds `if: <expr>` (forces object form even without `with:`)
- [ ] `--source <path>` flag sets `with: source: "<path>"`
- [ ] `--groups <list>` flag sets `with: groups: "<list>"`
- [ ] Generic `key=value` arguments set `with: key: "value"` entries
- [ ] If `pi.yaml` doesn't exist: prints "No pi.yaml found." and offers `pi init` flow (with confirm prompt); if confirmed, initializes the project and then adds the entry
- [ ] If the entry already exists verbatim: prints "already in pi.yaml" and exits 0 (idempotent)
- [ ] Prints a confirmation of what was added: `Added pi:install-python (version: 3.13) to setup.`
- [ ] When a short-form expansion happens, prints a note: `Resolved 'python' → pi:install-python`
- [ ] `go build ./...` and `go test ./...` pass

## Implementation Notes

### Command structure
`pi setup add <name> [flags] [key=value ...]`

This is a subcommand of the existing `pi setup` Cobra command. Add `setup add` as a subcommand in `internal/cli/setup.go` (or a new `internal/cli/setup_add.go`).

### Short-form resolution table
The resolution table lives in a `setupAddKnownTools` map:
```go
var setupAddKnownTools = map[string]string{
    // Short forms (no prefix)
    "python":    "pi:install-python",
    "node":      "pi:install-node",
    "nodejs":    "pi:install-node",
    "go":        "pi:install-go",
    "golang":    "pi:install-go",
    "rust":      "pi:install-rust",
    "ruby":      "pi:install-ruby",
    "uv":        "pi:install-uv",
    "tsx":       "pi:install-tsx",
    "homebrew":  "pi:install-homebrew",
    "brew":      "pi:install-homebrew",
    "terraform": "pi:install-terraform",
    "tf":        "pi:install-terraform",
    "kubectl":   "pi:install-kubectl",
    "k8s":       "pi:install-kubectl",
    "helm":      "pi:install-helm",
    "pnpm":      "pi:install-pnpm",
    "bun":       "pi:install-bun",
    "deno":      "pi:install-deno",
    "aws-cli":   "pi:install-aws-cli",
    "awscli":    "pi:install-aws-cli",
    "aws":       "pi:install-aws-cli",
    // pi: prefix short forms
    "pi:python":    "pi:install-python",
    "pi:node":      "pi:install-node",
    "pi:go":        "pi:install-go",
    "pi:rust":      "pi:install-rust",
    "pi:ruby":      "pi:install-ruby",
    "pi:uv":        "pi:install-uv",
    "pi:tsx":       "pi:install-tsx",
    "pi:homebrew":  "pi:install-homebrew",
    "pi:brew":      "pi:install-homebrew",
    "pi:terraform": "pi:install-terraform",
    "pi:kubectl":   "pi:install-kubectl",
    "pi:helm":      "pi:install-helm",
    "pi:pnpm":      "pi:install-pnpm",
    "pi:bun":       "pi:install-bun",
    "pi:deno":      "pi:install-deno",
    "pi:aws-cli":   "pi:install-aws-cli",
}
```

Resolution logic:
1. Check `setupAddKnownTools[name]` — if found, expand and print advisory
2. Otherwise, use the name as-is (could be a local automation like `setup/install-deps`, or a valid builtin like `pi:install-python`)

### YAML output rules
The logic for what YAML to write:

**Case 1: no flags** → bare string
```yaml
- pi:install-uv
```

**Case 2: `--version` only** → object form with `with:`
```yaml
- run: pi:install-python
  with:
    version: "3.13"
```

**Case 3: `--if` only** → object form with `if:`
```yaml
- run: pi:install-homebrew
  if: os.macos
```

**Case 4: `--if` + `--version`** → object form with both
```yaml
- run: pi:install-python
  if: os.macos
  with:
    version: "3.13"
```

**Case 5: local automation** → bare string (no `run:` wrapper)
```yaml
- setup/install-deps
```

**Case 6: local automation + `--if`** → object form
```yaml
- run: setup/configure-git-hooks
  if: dir.exists(".git")
```

**Rules for quoting `version:`**: always wrap the version value in double quotes to ensure YAML parses it as a string even when it looks like a float (e.g., `3.13` without quotes is a float).

### Writing to `pi.yaml`
Use the same line-based manipulation approach as `config.AddPackage()` in `internal/config/writer.go`. Add a new `AddSetupEntry(path string, entry SetupEntry) error` function in `writer.go` that:
1. Reads `pi.yaml`
2. Checks if the exact entry already exists (as a string match of the rendered YAML) → returns `DuplicateEntryError` if so
3. Finds the `setup:` block, locates the end of its list items, appends the new entry
4. If no `setup:` block exists, appends one at the end of the file
5. Re-validates the file with `config.Load()` after writing

The new `SetupEntry` rendering (for duplicate detection) must normalize the YAML consistently: same indentation, same key ordering, same quoting.

### No pi.yaml flow
```
No pi.yaml found. PI needs to be initialized before adding setup steps.

Initialize a new project in '<current dir>'? [Y/n] _
```
- If Y: run `initProject(root, inferredName)` (reuse from task 87), then proceed to add the entry
- If N: print "Aborted." and exit 1
- If `--yes` flag: skip the prompt and initialize

### Output examples

**Success (with expansion):**
```
Resolved 'python' → pi:install-python

Added to setup in pi.yaml:
  - run: pi:install-python
    with:
      version: "3.13"
```

**Success (no expansion):**
```
Added to setup in pi.yaml:
  - pi:install-uv
```

**Already exists:**
```
Already in pi.yaml. No changes made.
```

**No pi.yaml, accepted init:**
```
No pi.yaml found.

Initialize project 'vyper-platform'? [Y/n] y
Initialized project 'vyper-platform'.

Resolved 'python' → pi:install-python

Added to setup in pi.yaml:
  - run: pi:install-python
    with:
      version: "3.13"
```

### Flags
```
--version <v>         Sets with: version: "<v>"
--if <expr>           Sets if: <expr>
--source <path>       Sets with: source: "<path>"
--groups <list>       Sets with: groups: "<list>"
--yes / -y            Skip all prompts (use defaults / auto-confirm)
```

Additional `with:` entries come from positional `key=value` arguments after the name:
```bash
pi setup add pi:cursor/install-extensions file=.pi/cursor/extensions.txt
```

## Subtasks
- [ ] Add `setup add` subcommand in `internal/cli/setup_add.go`
- [ ] Implement `setupAddKnownTools` resolution table
- [ ] Implement YAML entry generation (all cases)
- [ ] Implement `AddSetupEntry()` in `internal/config/writer.go`
- [ ] Implement duplicate detection for setup entries
- [ ] Implement no-pi.yaml → init flow (reuse `initProject` from task 87)
- [ ] Wire `--version`, `--if`, `--source`, `--groups`, `--yes` flags
- [ ] Parse `key=value` positional arguments
- [ ] Print resolution advisory when short-form expansion happens
- [ ] Unit tests for `AddSetupEntry()`
- [ ] Integration tests in `tests/integration/setup_add_test.go`
- [ ] Update `docs/README.md` CLI reference table

## Blocked By
87-pi-init-command (needs `initProject` function from that task)
