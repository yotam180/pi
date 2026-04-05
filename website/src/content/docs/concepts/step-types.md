---
title: Step Types
description: The four step types — bash, python, typescript, and run — and how they work at runtime
---

PI supports four step types. Each can be used as inline code or a reference to an external file.

## In this section

- [`bash:`](#bash) — shell commands or `.sh` files
- [`python:`](#python) — Python scripts or `.py` files
- [`typescript:`](#typescript) — TypeScript scripts or `.ts` files (via tsx)
- [`run:`](#run) — call another automation
- [File paths vs inline](#file-paths-vs-inline) — how PI decides which you mean

---

## `bash:`

Run inline shell commands or a `.sh` file.

**Inline:**

```yaml
bash: echo "hello, world"
```

PI runs this via `bash -c "<script>" -- [args...]`. Arguments passed to the automation are available as `$1`, `$2`, etc.

**Multiline inline:**

```yaml
bash: |
  echo "Step 1: building..."
  go build -o bin/app ./...
  echo "Step 2: done"
```

**File reference:**

```yaml
bash: scripts/setup.sh
```

PI runs this via `bash <resolved_path> [args...]`. The file path is resolved relative to the automation YAML file's directory, not the project root.

Bash is the only step type that supports [`parent_shell: true`](/concepts/shell-shortcuts/#the-parent-shell-pattern).

## `python:`

Run inline Python or a `.py` file.

**Inline:**

```yaml
python: |
  import sys
  print(f"Hello from Python {sys.version}")
```

PI runs inline Python via `python3 -c "<script>" [args...]`. Arguments are available as `sys.argv[1:]`.

**File reference:**

```yaml
python: transform.py
```

PI runs file references via `python3 <resolved_path> [args...]`. The `.py` file path is relative to the automation YAML file — this means scripts live right next to their automation definition.

When a virtualenv is active (`$VIRTUAL_ENV` is set), PI uses `$VIRTUAL_ENV/bin/python` instead of `python3`.

## `typescript:`

Run inline TypeScript or a `.ts` file. Requires [`tsx`](https://github.com/privatenumber/tsx) to be installed.

**Inline:**

```yaml
typescript: |
  const greeting = "Hello from TypeScript";
  console.log(greeting);
```

PI writes inline TypeScript to a temporary file and runs it via `tsx <tmpfile> [args...]`. Arguments are available as `process.argv.slice(2)`.

**File reference:**

```yaml
typescript: process-data.ts
```

PI runs file references via `tsx <resolved_path> [args...]`. The `.ts` file path is relative to the automation YAML file.

If `tsx` is not in PATH, PI prints a clear error with an install hint: `npm install -g tsx`. You can also use the built-in installer: `pi run pi:install-tsx`.

## `run:`

Call another automation by name. This is how you compose automations.

**Local automation:**

```yaml
steps:
  - bash: echo "Building..."
  - run: docker/up
  - bash: echo "Environment ready"
```

PI resolves the name through the discovery system and recursively executes the target automation. Arguments are forwarded.

**Package automation (via alias):**

```yaml
steps:
  - run: mytools/docker/up
```

When packages are configured with aliases in `pi.yaml`, you can reference them with the alias prefix.

**On-demand GitHub reference:**

```yaml
steps:
  - run: org/repo@v1.0/docker/up
```

If the package isn't declared in `pi.yaml`, PI fetches it automatically from GitHub (once, then cached) and prints an advisory suggesting you add it to `packages:`.

**With inputs:**

```yaml
steps:
  - run: pi:install-python
    with:
      version: "3.13"
```

### What `run:` does NOT do

Each `run:` target is an independent automation execution:

- It does **not** inherit the caller's `env:` or automation-level `env:`
- It does **not** inherit the caller's `dir:`
- It does **not** inherit the caller's `timeout:`
- Circular dependencies (`a → b → a`) are detected and produce a clear error showing the full chain

## File paths vs inline

PI determines whether a value is a file path or inline code by checking the file extension:

- Ends in `.sh`, `.py`, or `.ts` → **file path** (resolved relative to the automation YAML file's directory)
- Anything else → **inline code**

```yaml
# File references
bash: scripts/setup.sh        # → runs the file
python: transform.py           # → runs the file
typescript: process.ts         # → runs the file

# Inline code
bash: echo "hello"             # → inline
python: print("hello")         # → inline
typescript: console.log("hi")  # → inline
```

This convention lets automation assets live right next to the YAML file that uses them:

```
.pi/
  docker/
    up.yaml
    logs-formatted.yaml
    logs-formatted.py          ← script lives next to its automation
  setup/
    install-cursor-extensions/
      automation.yaml
      extensions.txt           ← bundled asset
```
