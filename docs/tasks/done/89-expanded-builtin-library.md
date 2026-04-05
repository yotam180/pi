# Expanded Built-in Library

## Type
feature

## Status
todo

## Priority
medium

## Project
standalone

## Description
Expand the built-in automation library to cover the full range of tools a modern development project needs. This task adds new installer automations (terraform, kubectl, helm, pnpm, bun, deno, aws-cli), two new utility automations (pi:uv/sync and pi:set-env), and a pi:npm/install builtin. All new installers must follow the exact same pattern as the existing ones — `install:` block, `version` input, `first:` block for trying mise before platform-specific managers, consistent status output.

The source of truth for the existing pattern is `internal/builtins/embed_pi/install-python.yaml`. New installers must match it structurally.

Read `docs/philosophy.md` before implementing. Consistency (principle 6) is the governing concern here.

## Acceptance Criteria

### New installer automations (all in `internal/builtins/embed_pi/`)
- [ ] `install-terraform.yaml` — installs Terraform via mise/brew/tfenv
- [ ] `install-kubectl.yaml` — installs kubectl via mise/brew
- [ ] `install-helm.yaml` — installs Helm via brew/script
- [ ] `install-pnpm.yaml` — installs pnpm via npm/script
- [ ] `install-bun.yaml` — installs Bun via the official install script
- [ ] `install-deno.yaml` — installs Deno via the official install script
- [ ] `install-aws-cli.yaml` — installs AWS CLI v2 via the official installer
- [ ] All new installers: have a `version` input (optional where applicable), a `test:` phase, a `run:` phase with `first:` blocks prioritizing mise, and a `version:` command

### New utility automations
- [ ] `uv/sync.yaml` — runs `uv sync` with optional `groups` and `args` inputs; uses `install:` lifecycle to check if already synced
- [ ] `set-env.yaml` — idempotently writes `export KEY=VALUE` to `~/.zshrc` and `~/.bashrc`
- [ ] `npm/install.yaml` — runs `npm ci` (preferred) or `npm install` in an optional directory

### Builtins test coverage
- [ ] `builtins_test.go` updated: 7 new installer tests + 3 new utility tests
- [ ] All new automations: `pi list --builtins` shows them; `pi info pi:<name>` shows correct description and inputs
- [ ] `go build ./...` and `go test ./...` pass

## Implementation Notes

Read `docs/architecture.md` → "Built-in automations" section before starting. The pattern is well-established there.

All YAML files go in `internal/builtins/embed_pi/`. The `//go:embed` directive in `builtins.go` picks them up automatically.

---

### Pattern reference: `install-python.yaml`
All new installers must follow this exact structural pattern:
```yaml
description: Install Python at a specific version

inputs:
  version:
    type: string
    description: Python version to install (e.g. "3.13")

install:
  test:
    - bash: python3 --version 2>&1 | grep -q "Python $PI_IN_VERSION"
  run:
    - first:
        - bash: mise install "python@$PI_IN_VERSION" && mise use -g "python@$PI_IN_VERSION"
          if: command.mise
        - bash: brew install "python@$PI_IN_VERSION"
          if: command.brew
        - bash: |
            echo "no suitable installer found (tried mise, brew)" >&2
            exit 1
  version: python3 --version 2>&1 | awk '{print $2}'
```

When `version` input is optional (tool doesn't require a specific version): make it optional with an empty default and handle the no-version case in the bash commands.

---

### `install-terraform.yaml`
```yaml
description: Install Terraform at a specific version

inputs:
  version:
    type: string
    description: Terraform version to install (e.g. "1.7")

install:
  test:
    - bash: terraform version 2>&1 | grep -q "Terraform v$PI_IN_VERSION"
  run:
    - first:
        - bash: mise install "terraform@$PI_IN_VERSION" && mise use -g "terraform@$PI_IN_VERSION"
          if: command.mise
        - bash: brew install "hashicorp/tap/terraform@$PI_IN_VERSION" || brew install terraform
          if: command.brew
        - bash: |
            echo "no suitable installer found (tried mise, brew)" >&2
            echo "Install manually: https://developer.hashicorp.com/terraform/install" >&2
            exit 1
  version: terraform version 2>&1 | head -1 | awk '{print $2}' | tr -d 'v'
```

---

### `install-kubectl.yaml`
```yaml
description: Install kubectl at a specific version

inputs:
  version:
    type: string
    description: kubectl version to install (e.g. "1.28")

install:
  test:
    - bash: kubectl version --client 2>&1 | grep -q "v$PI_IN_VERSION"
  run:
    - first:
        - bash: mise install "kubectl@$PI_IN_VERSION" && mise use -g "kubectl@$PI_IN_VERSION"
          if: command.mise
        - bash: brew install kubectl
          if: command.brew
        - bash: |
            echo "no suitable installer found (tried mise, brew)" >&2
            echo "Install manually: https://kubernetes.io/docs/tasks/tools/" >&2
            exit 1
  version: kubectl version --client 2>&1 | grep -o 'v[0-9]*\.[0-9]*\.[0-9]*' | head -1 | tr -d 'v'
```

---

### `install-helm.yaml`
```yaml
description: Install Helm at a specific version

inputs:
  version:
    type: string
    description: Helm version to install (e.g. "3.14")

install:
  test:
    - bash: helm version 2>&1 | grep -q "v$PI_IN_VERSION"
  run:
    - first:
        - bash: mise install "helm@$PI_IN_VERSION" && mise use -g "helm@$PI_IN_VERSION"
          if: command.mise
        - bash: brew install helm
          if: command.brew
        - bash: curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
  version: helm version --short 2>&1 | grep -o 'v[0-9]*\.[0-9]*\.[0-9]*' | tr -d 'v'
```

---

### `install-pnpm.yaml`
```yaml
description: Install pnpm

inputs:
  version:
    type: string
    description: pnpm version to install (optional; omit for latest)

install:
  test:
    - bash: command -v pnpm >/dev/null 2>&1
  run:
    - first:
        - bash: mise install "pnpm@${PI_IN_VERSION:-latest}" && mise use -g "pnpm@${PI_IN_VERSION:-latest}"
          if: command.mise
        - bash: npm install -g pnpm${PI_IN_VERSION:+@$PI_IN_VERSION}
          if: command.npm
        - bash: curl -fsSL https://get.pnpm.io/install.sh | sh -
  version: pnpm --version
```

---

### `install-bun.yaml`
```yaml
description: Install Bun JavaScript runtime

inputs:
  version:
    type: string
    description: Bun version to install (optional; omit for latest)

install:
  test:
    - bash: command -v bun >/dev/null 2>&1
  run:
    - first:
        - bash: mise install "bun@${PI_IN_VERSION:-latest}" && mise use -g "bun@${PI_IN_VERSION:-latest}"
          if: command.mise
        - bash: brew install bun
          if: command.brew
        - bash: curl -fsSL https://bun.sh/install | bash
  version: bun --version
```

---

### `install-deno.yaml`
```yaml
description: Install Deno JavaScript/TypeScript runtime

inputs:
  version:
    type: string
    description: Deno version to install (optional; omit for latest)

install:
  test:
    - bash: command -v deno >/dev/null 2>&1
  run:
    - first:
        - bash: mise install "deno@${PI_IN_VERSION:-latest}" && mise use -g "deno@${PI_IN_VERSION:-latest}"
          if: command.mise
        - bash: brew install deno
          if: command.brew
        - bash: curl -fsSL https://deno.land/install.sh | sh
  version: deno --version 2>&1 | head -1 | awk '{print $2}'
```

---

### `install-aws-cli.yaml`
```yaml
description: Install AWS CLI v2

install:
  test:
    - bash: aws --version 2>&1 | grep -q "aws-cli/2"
  run:
    - first:
        - bash: brew install awscli
          if: command.brew and os.macos
        - bash: |
            curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o /tmp/awscliv2.zip
            unzip -q /tmp/awscliv2.zip -d /tmp
            sudo /tmp/aws/install
            rm -rf /tmp/aws /tmp/awscliv2.zip
          if: os.linux and os.arch.amd64
        - bash: |
            curl "https://awscli.amazonaws.com/awscli-exe-linux-aarch64.zip" -o /tmp/awscliv2.zip
            unzip -q /tmp/awscliv2.zip -d /tmp
            sudo /tmp/aws/install
            rm -rf /tmp/aws /tmp/awscliv2.zip
          if: os.linux and os.arch.arm64
        - bash: |
            echo "AWS CLI v2 install not supported on this platform" >&2
            echo "See: https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html" >&2
            exit 1
  version: aws --version 2>&1 | awk '{print $1}' | cut -d'/' -f2
```

---

### `uv/sync.yaml`
This is a utility, not an installer. Uses `steps:` not `install:`.
```yaml
description: Sync Python project dependencies using uv

inputs:
  dir:
    type: string
    description: Directory containing pyproject.toml (default: project root)
  groups:
    type: string
    description: Dependency groups to include (e.g. "dev,local")
  args:
    type: string
    description: Extra arguments passed to uv sync

bash: |
  SYNC_DIR="${PI_IN_DIR:-.}"
  GROUPS=""
  if [ -n "$PI_IN_GROUPS" ]; then
    for g in $(echo "$PI_IN_GROUPS" | tr ',' ' '); do
      GROUPS="$GROUPS --group $g"
    done
  fi
  cd "$SYNC_DIR"
  uv sync $GROUPS $PI_IN_ARGS
```

Note: uses single-step shorthand with `bash:` at top level. The `dir:` step modifier is not used here because we need to compute the path from the input — instead we `cd` inside the script.

---

### `set-env.yaml`
Idempotently writes an environment variable to the user's shell rc files. Does NOT use `parent_shell: true` — it writes to the rc file directly. The `pi-setup-<project>` helper auto-sources the rc file after setup completes, so the var is available immediately.

```yaml
description: Idempotently add an environment variable export to the shell config

inputs:
  key:
    type: string
    description: Environment variable name (e.g. "VYPER_HOME")
  value:
    type: string
    description: Value to export (e.g. "/Users/me/projects/myapp")
  comment:
    type: string
    description: Optional comment to add above the export line

bash: |
  set_env_if_missing() {
    local rc_file="$1"
    local key="$PI_IN_KEY"
    local value="$PI_IN_VALUE"
    local comment="$PI_IN_COMMENT"

    [ -f "$rc_file" ] || return 0

    if grep -q "export $key=" "$rc_file"; then
      echo "  $key already set in $rc_file"
    else
      if [ -n "$comment" ]; then
        echo "# $comment" >> "$rc_file"
      fi
      echo "export $key=\"$value\"" >> "$rc_file"
      echo "  added $key to $rc_file"
    fi
  }

  set_env_if_missing "$HOME/.zshrc"
  set_env_if_missing "$HOME/.bashrc"
```

---

### `npm/install.yaml`
```yaml
description: Install Node.js dependencies using npm ci (or npm install as fallback)

inputs:
  dir:
    type: string
    description: Directory containing package.json (default: project root)

bash: |
  TARGET="${PI_IN_DIR:-.}"
  cd "$TARGET"
  if [ -f "package-lock.json" ]; then
    npm ci
  else
    npm install
  fi
```

---

### Testing
New tests go in `internal/builtins/builtins_test.go`. For each new builtin, add:
1. A test that verifies the automation is discovered with the correct name (e.g., `pi:install-terraform`)
2. A test that verifies the `inputs:` block is parsed correctly for installers with inputs
3. A test that verifies the `if:` conditions and `first:` block structure for the installer's `run:` phase

Follow the pattern of the existing installer tests in that file.

## Subtasks
- [x] Create `install-terraform.yaml`
- [x] Create `install-kubectl.yaml`
- [x] Create `install-helm.yaml`
- [x] Create `install-pnpm.yaml`
- [x] Create `install-bun.yaml`
- [x] Create `install-deno.yaml`
- [x] Create `install-aws-cli.yaml`
- [x] Create `uv/sync.yaml` (in `embed_pi/uv/`)
- [x] Create `set-env.yaml`
- [x] Create `npm/install.yaml` (in `embed_pi/npm/`)
- [x] Update `builtins_test.go` with tests for all new automations
- [x] Update `docs/README.md` Built-in library table
- [x] Verify `pi list --builtins` shows all new automations
- [x] Verify `pi info pi:install-terraform` etc. shows correct descriptions

## Blocked By
None (can proceed independently of tasks 87 and 88)

## Implementation Notes

### Session 2026-04-05

All 10 new automations created and tested:

**Installer automations (7):**
- `install-terraform`, `install-kubectl`, `install-helm` — required `version` input, mise→brew→error fallback chain via `first:` blocks
- `install-pnpm` — optional `version` input, mise→npm→curl fallback chain
- `install-bun`, `install-deno` — optional `version` input, mise→brew→curl fallback chain
- `install-aws-cli` — no version input, brew on macOS→official zip on Linux (amd64/arm64)→error

**Utility automations (3):**
- `uv/sync` — all inputs optional (`dir`, `groups`, `args`), uses single-step shorthand
- `set-env` — `key` and `value` required, `comment` optional, writes to both `.zshrc` and `.bashrc`
- `npm/install` — `dir` optional, prefers `npm ci` when `package-lock.json` exists

**Key decisions:**
- Used `required: false` for optional inputs (not `default: ""`) because `IsRequired()` treats empty string default as "no default"
- YAML descriptions containing colons need quoting (e.g. `"Directory containing package.json (default: project root)"`)
- `setupAddKnownTools` map in `setup_add.go` already had entries for all new tools (terraform, kubectl, helm, pnpm, bun, deno, aws-cli)
- 55 new test cases added to `builtins_test.go` (total now 136 subtests)
- Updated `docs/README.md` built-in library list and `docs/architecture.md` package structure + design decisions
