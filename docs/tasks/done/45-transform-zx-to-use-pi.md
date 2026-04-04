# Transform zx to Use PI

## Type
feature

## Status
done

## Priority
medium

## Project
10-zx-adoption-test

## Description
Create PI automations for zx's developer workflows in `~/projects/zx`. This is the final phase of the adoption test — actually using PI to model a real TypeScript project's workflows.

Follow the playbook at `docs/playbooks/real-world-adoption-test.md` (Phase 3).

### Steps
1. Build the development `pi` binary: `cd ~/projects/vyper-tooling && go build -o ~/go/bin/pi ./cmd/pi/`
2. Create `~/projects/zx/.pi/` folder
3. Create `~/projects/zx/pi.yaml` with project config
4. Write automation YAML files for each workflow from task 43:
   - Build/compile
   - Test: unit tests, integration tests
   - Lint/format
   - Setup: Node.js, npm install
5. Define setup automations for dev environment
6. Define shell shortcuts for common operations
7. Test every automation against the original commands
8. Test `pi setup` and `pi shell`
9. Document any quirks and whether they should be solved in YAML or as new PI features

## Acceptance Criteria
- [x] `~/projects/zx/.pi/` contains automations for all major workflows
- [x] `~/projects/zx/pi.yaml` has shortcuts and setup entries
- [x] `pi setup` in zx installs all required tools
- [x] `pi shell` in zx installs working shortcuts
- [x] `pi run <name>` for each automation produces correct results
- [x] `pi list` shows all automations with descriptions
- [x] Quirks documented in Implementation Notes with feature-vs-YAML analysis

## Implementation Notes

### Automations Created (17 local)

| Automation | Command | Tested |
|-----------|---------|--------|
| `build` | `npm run build` | ✓ Full build succeeded |
| `test` | `npm test` | ✓ Full suite runs |
| `test/unit` | `npm run test:unit` | ✓ 267 tests, 261 pass (5 pre-existing failures in log color tests) |
| `test/coverage` | `npm run test:coverage` | ✓ |
| `test/types` | `npm run test:types` | ✓ |
| `test/size` | `npm run test:size` | ✓ All bundles within limits |
| `test/circular` | `npm run test:circular` | ✓ No circular dependencies |
| `test/license` | `npm run test:license` | ✓ 1/1 pass |
| `test/audit` | `npm run test:audit` | ✓ |
| `fmt` | `npm run fmt` | ✓ Formatted pi.yaml |
| `fmt/check` | `npm run fmt:check` | ✓ All files pass |
| `docs/dev` | `npm run docs:dev` | ✓ (VitePress dev server) |
| `docs/build` | `npm run docs:build` | ✓ |
| `docs/preview` | `npm run docs:preview` | ✓ |
| `docker/build` | `npm run build:dcr` | ✓ |
| `setup/check-node` | Verify Node.js is available | ✓ Detected v25.9.0 |
| `setup/install-deps` | `npm ci` | ✓ Idempotent (lockfile v3) |

### Shortcuts (8)

| Shortcut | Automation |
|----------|-----------|
| `zxb` | build |
| `zxt` | test |
| `zxtu` | test/unit |
| `zxtc` | test/coverage |
| `zxf` | fmt |
| `zxfc` | fmt/check |
| `zxdd` | docs/dev |
| `zxdb` | docs/build |

### Setup

`pi setup` runs:
1. `setup/check-node` — verifies Node.js is installed, shows version
2. `setup/install-deps` — runs `npm ci` if `node_modules/` doesn't exist

Both are installer automations with test/run/version lifecycle, producing clean PI status output.

### pi doctor

All 15 automations with `requires:` show ✓ for their requirements.

### Quirks & Decisions

1. **Used `setup/check-node` instead of `pi:install-node`**: The built-in `pi:install-node` checks for exact major version match and tries to install via mise/brew. For zx, any Node >= 20 works, so a simple check-and-report is more appropriate. This is by design — `pi:install-node` is for projects that need a specific version.

2. **Pre-existing test failures**: 5 unit tests in `test/log.test.ts` fail due to ANSI color code expectations vs. terminal detection. These fail identically whether run via `pi run test/unit` or `npm run test:unit` — not PI-related.

3. **Prettier formats pi.yaml**: The project uses Prettier with `--write .`, which picks up `pi.yaml`. This is expected and desirable — PI's YAML is valid YAML and Prettier handles it correctly.

4. **All workflows are npm wrappers**: Every automation is a single `bash: npm run <script>` step. This matches the httpie pattern (single `bash: python -m pytest` etc.). PI's value is in the orchestration layer: shortcuts, setup, discovery, doctor, structured info.

5. **Zero feature gaps**: Same conclusion as tasks 43 and 44. PI fully supports TypeScript/Node.js projects.

### Comparison with Previous Adoption Tests

| Aspect | fzf (Go) | bat (Rust) | httpie (Python) | zx (TypeScript) |
|--------|----------|-----------|-----------------|-----------------|
| Automations | 11 | 10 | 10 | 17 |
| Shortcuts | 6 | 6 | 7 | 8 |
| Feature gaps | 2 (env:, install-go) | 1 (install-rust) | 0 | 0 |
| Setup entries | 3 | 3 | 3 | 2 |
| Primary tool | make/go | cargo | pip/pytest | npm |

zx has the most automations because it has the most granular test suite (7 test automations vs. 2-3 in other projects).

## Subtasks
- [x] Create `.pi/` and `pi.yaml`
- [x] Write build automations
- [x] Write test automations
- [x] Write lint/format automations
- [x] Write setup automations
- [x] Define shortcuts
- [x] Test all automations
- [x] Test `pi setup` and `pi shell`
- [x] Document quirks and lessons learned

## Blocked By
43-clone-and-examine-zx
44-implement-zx-feature-gaps
