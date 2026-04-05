# Add a command to scaffold new automations

## Type
feature

## Status
done

## Priority
high

## Project
standalone

## Description
There is no way for a new user to discover the automation YAML schema without trial and error. When `pi init` creates an empty `.pi/` directory, the user has no examples, templates, or schema reference to work from.

I tried three incorrect formats before landing on the right one:
1. `steps: [{ run: cargo build }]` → error (run references automations, not shell commands)
2. `steps: [{ cmd: cargo build }]` → error (cmd is not a valid step type)
3. Only after the error message told me "must specify one of: bash, run, python, typescript, first" did I know what to use.

### Proposed Solution
Add a `pi new` (or `pi create`) command that scaffolds an automation file:

```
pi new build
# Creates .pi/build.yaml with a working template:
#   name: build
#   description: TODO
#   steps:
#     - bash: echo "hello world"
```

Advanced usage:
```
pi new setup/install-deps    # Creates .pi/setup/install-deps.yaml
pi new build --bash "cargo build"   # Pre-fills the command
```

### Alternative / Complementary
Even without a `pi new` command, `pi init` should create a commented example file in `.pi/` like `.pi/_example.yaml.sample` showing the full schema with comments.

## Acceptance Criteria
- [x] A new user can create their first automation without guessing the YAML format
- [x] Either `pi new` exists, or `pi init` includes a reference/example file

## Implementation Notes

### Approach
Implemented both solutions from the acceptance criteria:

1. **`pi new` command** — scaffolds a new automation YAML file at `.pi/<name>.yaml`
   - `pi new build` creates `.pi/build.yaml` with a default bash echo
   - `pi new setup/install-deps` creates nested directories automatically
   - `--bash "cmd"` pre-fills with a bash command
   - `--python "script.py"` pre-fills with a python script reference
   - `--description "text"` or `-d "text"` sets the description
   - Strips `.yaml`/`.yml` extension if user accidentally includes it
   - Refuses to overwrite existing files (idempotent safety)
   - Suggests `pi init` if no project exists

2. **`pi init` now creates `.pi/hello.yaml`** — a working example automation
   - Content: `description: A sample automation — edit or delete this file\nbash: echo "Hello from PI!"`
   - Next steps output updated to show `pi run hello` and `pi new build --bash "..."`

### File layout
- `internal/cli/new.go` — command, `runNew()`, `findPiDir()`, `generateAutomationYAML()`, `ExampleAutomationContent`
- `internal/cli/new_test.go` — 14 tests (basic, bash flag, python flag, description, nested path, already exists, no project, strip extension, output, valid content, generate helpers, init creates example, init next steps)
- `internal/cli/init.go` — updated `initProject()` and `printNextSteps()`
- `internal/cli/root.go` — registered `newNewCmd()`

### Design decisions
- Uses single-step shorthand (top-level `bash:` / `python:`) per philosophy principle #3 (short form is the real form)
- Default template includes `description: TODO` placeholder to encourage documentation
- Generated files are valid YAML that `pi validate` and `pi run` accept immediately
- No `name:` field in generated files — PI derives it from the path (per docs/README.md)

## Subtasks
- [x] Implement pi new command with --bash, --python, --description flags
- [x] Update pi init to create .pi/hello.yaml example
- [x] Update pi init next steps to mention pi run hello and pi new
- [x] 14 tests for pi new + 2 tests for init changes

## Blocked By
