# `pi init` Command

## Type
feature

## Status
done

## Priority
high

## Project
standalone

## Description
Add a `pi init` command that bootstraps a new PI project: creates `pi.yaml` with the project name and creates the `.pi/` directory. The command is interactive by default — it infers the project name from the directory, asks the developer to confirm or change it, and explains what it created. It also handles being run in a project that is already initialized gracefully.

Read `docs/philosophy.md` before implementing. This command is the very first thing a new developer touches — it must reflect every principle there.

## Acceptance Criteria
- [ ] `pi init` creates `pi.yaml` with `project: <name>` and creates `.pi/`
- [ ] Project name defaults to the current directory name (kebab-cased); developer can accept or override via prompt
- [ ] `pi init --name <name>` skips the prompt entirely (non-interactive mode)
- [ ] `pi init --yes` skips the prompt and uses the inferred name (for scripts)
- [ ] If `pi.yaml` already exists, prints "Already initialized (project: <name>)" and suggests the next step (`pi setup add`, `pi shell`)
- [ ] If `.pi/` already exists but `pi.yaml` does not, creates only `pi.yaml` (no error)
- [ ] After successful init, prints a concise "what's next" guide:
  - Add setup steps: `pi setup add pi:python --version 3.13`
  - Or run directly: `pi run <name>`
  - Install shell shortcuts: `pi shell`
- [ ] `go build ./...` and `go test ./...` pass
- [ ] Integration test: `pi init` in a temp dir creates the expected files

## Implementation Notes

### Decisions made
- `initProject(root, name, stdout)` is the exported reusable function — accepts an `io.Writer` for stdout so `setup add` can call it and capture output
- `promptProjectName()` is a separate function from `resolveProjectName()` for testability — tests can call the prompt logic directly without needing a real terminal
- Non-interactive detection uses `os.ModeCharDevice` check on stdin, which correctly detects piped input (CI, scripts, `exec.Command` in tests)
- The "Next steps" block is shown on both success and already-initialized paths — the developer always sees what to do next
### Command location
`internal/cli/init.go` — new file. Register the command in `root.go`.

### Prompt behavior
Use `fmt.Scan` or `bufio.NewReader(os.Stdin).ReadString('\n')` for the project name prompt. No external readline library — keep it simple.

Prompt format:
```
Initializing PI project.

Project name [my-project]: _
```
The default (in brackets) is the current directory name, converted to kebab-case.

If the user presses Enter without typing, use the default. If they type a name, use that.

### Kebab-case conversion for directory name
Convert the directory name: lowercase, replace spaces and underscores with `-`, strip characters that aren't `a-z`, `0-9`, or `-`.

### What `pi init` creates

1. `pi.yaml`:
```yaml
project: <name>
```
That's it. No placeholder shortcuts or setup entries. Keep it minimal — the developer adds what they need.

2. `.pi/` directory:
```
.pi/   (empty directory)
```

### Non-interactive mode
- `--name <name>`: use this name, no prompt, print "Initialized project '<name>'"
- `--yes` / `-y`: accept the inferred name, no prompt
- If stdin is not a terminal (piped), behave as `--yes` (never block on a prompt in CI)

### Already-initialized handling
If `pi.yaml` exists:
```
Already initialized (project: my-project).

Next steps:
  pi setup add pi:python --version 3.13   add a setup step
  pi shell                                 install shell shortcuts
  pi run <name>                            run an automation
```
Exit 0. This is not an error.

### What's-next message (success case)
```
Initialized project 'my-project'.

  Created pi.yaml
  Created .pi/

Next steps:
  pi setup add pi:python --version 3.13   add a setup step
  pi shell                                 install shell shortcuts
  pi run <name>                            run an automation
```

The "Next steps" block uses the same dim-style output that other PI commands use (`display.Printer.Dim()`).

### CLI signature
```
pi init [--name <name>] [--yes]
```

No positional arguments. The project name comes from `--name` or the interactive prompt.

### Integration with `pi setup add`
`pi setup add` calls this flow if `pi.yaml` is missing (see task 88). The init flow should be callable as a function from `add.go`, not just from the CLI. Extract the core logic into a function `initProject(root, name string) error` in `init.go` so `add.go` can reuse it.

## Subtasks
- [x] Create `internal/cli/init.go` with the `pi init` command
- [x] Implement interactive prompt with inferred default
- [x] Implement `--name` and `--yes` flags
- [x] Handle already-initialized case gracefully
- [x] Print "what's next" message on success
- [x] Register command in `root.go`
- [x] Add unit tests (13 tests in `init_test.go`)
- [x] Add integration test in `tests/integration/init_test.go` (8 tests)
- [x] Update `docs/README.md` CLI reference table with `pi init`

## Blocked By
None
