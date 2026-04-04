# Clone and Examine zx Workflows

## Type
research

## Status
in_progress

## Priority
high

## Project
10-zx-adoption-test

## Description
Clone google/zx into `~/projects/zx` and examine all developer workflows. Document every build command, test command, lint/format command, CI workflow, and release process. For each workflow, assess whether PI can model it today or if a new feature is needed.

Follow the playbook at `docs/playbooks/real-world-adoption-test.md` (Phase 1).

### Steps
1. `git clone https://github.com/google/zx.git ~/projects/zx`
2. Read `package.json` — document build config, scripts, and dependencies
3. Read `tsconfig.json` and any build configuration
4. Read CI workflows (`.github/workflows/`)
5. Read any release configuration
6. List all tools/runtimes required (Node.js, npm, TypeScript, etc.)
7. For each workflow, note whether PI can model it and what's missing
8. Record all findings in Implementation Notes

## Acceptance Criteria
- [x] zx cloned to `~/projects/zx`
- [x] Every build/test/lint command documented
- [x] Every CI workflow documented
- [x] Required tools/runtimes listed
- [x] PI feature gap analysis completed
- [x] Findings recorded in Implementation Notes

## Implementation Notes

### Repo Overview
zx v8.9.0 — "A tool for writing better scripts." TypeScript/Node.js project (~43k stars). Apache-2.0 license. Uses npm (with lockfile), esbuild for JS bundling, tsc for declaration files, and extensive custom build scripts in `scripts/`.

### Required Tools/Runtimes
| Tool | Version | Purpose |
|------|---------|---------|
| Node.js | 24 (`.node_version`) | Runtime, build, test |
| npm | (bundled with node) | Package manager |
| TypeScript | 5.9.3 (devDep) | Type checking, declaration generation |
| esbuild | 0.27.3 (devDep) | JS bundling |
| tsx | 4.21.0 (devDep, local) | TS execution for smoke tests |
| c8 | 11.0.0 (devDep) | Code coverage |
| prettier | 3.8.1 (devDep) | Formatting |
| tsd | 0.33.0 (devDep) | Type definition testing |
| size-limit | 12.0.0 (devDep) | Bundle size checking |
| madge | 8.0.0 (devDep) | Circular dependency detection |
| lefthook | 2.1.1 (devDep) | Git hooks manager |
| commitlint | ^20.4.2 (devDep) | Conventional commit enforcement |
| Docker | any | Docker image build/test |
| vitepress | 1.6.4 (devDep) | Documentation site |
| zizmor | 1.22.0 (external, via uvx) | CI workflow security scanner |

### Developer Workflows (from `package.json` scripts)

#### Build Workflows
| Command | What it does | PI can model? |
|---------|-------------|---------------|
| `npm run build` | Full build: versions → JS → DTS → tests → clean → manifests | Yes (single bash step) |
| `npm run build:js` | esbuild JS bundling (CJS+ESM hybrid) | Yes (single bash step) |
| `npm run build:vendor` | Bundle vendor dependencies | Yes |
| `npm run build:versions` | Generate version constants | Yes |
| `npm run build:tests` | Build test fixtures | Yes |
| `npm run build:dts` | tsc declarations + post-processing | Yes |
| `npm run build:dcr` | Docker image build | Yes |
| `npm run build:jsr` | JSR package manifest | Yes |
| `npm run build:lite` | Lite package.json variant | Yes |
| `npm run build:pkgjson` | Main package.json variant | Yes |
| `npm run build:manifest` | Combined manifest builds | Yes |

#### Test Workflows
| Command | What it does | PI can model? |
|---------|-------------|---------------|
| `npm test` | Full suite: size → fmt:check → unit → types → license | Yes |
| `npm run test:unit` | `node --experimental-transform-types ./test/all.test.js` | Yes |
| `npm run test:coverage` | c8 coverage + threshold checks | Yes |
| `npm run test:types` | tsd type tests | Yes |
| `npm run test:license` | License compliance check | Yes |
| `npm run test:size` | Bundle size limit check | Yes |
| `npm run test:circular` | Circular dependency check via madge | Yes |
| `npm run test:audit` | `npm audit` | Yes |
| `npm run test:npm` | npm publish dry-run test | Yes |
| `npm run test:jsr` | JSR publish dry-run test | Yes |
| `npm run test:dcr` | Docker container runtime test | Yes |
| `npm run test:workflow` | GitHub Actions security scan (zizmor) | Yes |
| `npm run test:smoke:*` | Various smoke tests (tsx, tsc, ts-node, bun, deno, win32, cjs, mjs) | Yes |

#### Format/Lint Workflows
| Command | What it does | PI can model? |
|---------|-------------|---------------|
| `npm run fmt` | `prettier --write .` | Yes |
| `npm run fmt:check` | `prettier --check .` (CI mode) | Yes |

#### Docs Workflows
| Command | What it does | PI can model? |
|---------|-------------|---------------|
| `npm run docs:dev` | VitePress dev server | Yes |
| `npm run docs:build` | VitePress static build | Yes |
| `npm run docs:preview` | VitePress preview server | Yes |

### Git Hooks (lefthook.yml)
| Hook | Commands | PI can model? |
|------|----------|---------------|
| pre-commit | `npm run fmt && git add {staged_files}` (parallel, glob: `*.{js,ts,md,yml,yaml}`) | Yes (but PI doesn't manage git hooks natively — lefthook is a separate tool) |
| commit-msg | `npx commitlint --edit` | Yes |
| pre-push | `test:license`, `test:size`, `test:circular` (parallel) | Yes |

### CI Workflows (.github/workflows/)

#### 1. test.yml — Main Test Pipeline
- **Trigger**: push, pull_request, schedule (every 4 days)
- **Jobs**:
  - `build` (ubuntu, Node 24): `npm ci` → `npm run build` → upload artifact
  - `checks` (ubuntu, Node 24): download build → fmt:check, license, size, audit, circular, npm bundle test, JSR dry-run, commitlint (PR only)
  - `test` (ubuntu, Node 24): download build → test:coverage, test:types
  - `docker-test`: build Docker image + run test:dcr
  - `smoke-win32-node16`: Windows matrix (2022, 2025), Node 16
  - `smoke-bun`: Bun runtime tests
  - `smoke-deno`: Deno 1 & 2 matrix
  - `smoke-node`: Node version matrix (12, 14, 16, 18, 20, 22, 24, 25-nightly)
  - `smoke-graal`: GraalVM 17 & 20
  - `smoke-ts`: TypeScript version matrix (4, 5, rc, next)

#### 2. publish.yml — Release Pipeline
- **Trigger**: release created, workflow_dispatch
- Build → test → publish to Google npm, GitHub npm, JSR, Docker (ghcr.io)
- Version tag validation against package.json

#### 3. dev-publish.yml — Dev Release
- **Trigger**: workflow_dispatch
- Same as publish but with `-dev.SHA` version suffix, `dev` tag

#### 4. docs.yml — Documentation Deployment
- **Trigger**: release created, workflow_dispatch
- VitePress build → GitHub Pages

#### 5. jsr-publish.yml — Manual JSR Publish
- **Trigger**: workflow_dispatch
- Build → test → publish to jsr.io

#### 6. codeql.yml — Security Analysis
- **Trigger**: push/PR to main, weekly schedule
- GitHub CodeQL for JavaScript/TypeScript

#### 7. osv.yml — Vulnerability Scanning
- **Trigger**: push/PR to main, weekly schedule
- Google OSV-Scanner

#### 8. zizmor.yml — CI Workflow Linting
- **Trigger**: push to main, all PRs
- zizmor via uvx for GitHub Actions security

### PI Feature Gap Analysis

**All zx workflows can be modeled with PI today.** Every workflow is either:
1. A simple `npm run <script>` command → single bash step
2. A Docker command → single bash step
3. A setup step (`npm ci`) → single bash step

There are **zero feature gaps**. The key observations:

1. **npm-based workflow**: All workflows delegate to npm scripts. PI wraps these as `bash: npm run <script>` steps. This is the correct level of abstraction — PI should orchestrate npm, not replace it.

2. **No new step types needed**: Everything runs through npm/node. PI's existing bash steps handle all of these.

3. **Setup is straightforward**: Just Node.js + `npm ci`. PI's `pi:install-node` built-in covers the Node.js installation, and `npm ci` is a bash step.

4. **lefthook manages git hooks**: The project uses lefthook (npm devDep) for git hooks. PI's `pi:git/install-hooks` is designed for static hook scripts, not lefthook. This is fine — PI's setup automation can ensure lefthook is installed via `npm ci`, and lefthook manages hooks itself.

5. **No new built-in automations needed**: `pi:install-node` already handles Node.js. `npm ci` handles all dependencies. No need for a dedicated `pi:install-npm` since npm comes with Node.js.

6. **Multi-runtime smoke tests** (Bun, Deno, GraalVM) are CI-only and don't need PI automations — they're not developer workflows.

### Summary
This is the cleanest adoption test so far. Like httpie (Python project where everything goes through pip/pytest), zx is an npm project where everything goes through npm scripts. PI acts as the orchestrator layer — providing shortcuts, setup, and structured automation YAML — while the actual tool chain is npm + esbuild + tsc.

**Recommended PI automations for zx:**
- `setup/install-deps` → `npm ci`
- `build` → `npm run build`
- `test` → `npm test`
- `test/unit` → `npm run test:unit`
- `test/coverage` → `npm run test:coverage`
- `test/types` → `npm run test:types`
- `fmt` → `npm run fmt`
- `fmt/check` → `npm run fmt:check`
- `docs/dev` → `npm run docs:dev`
- `docs/build` → `npm run docs:build`
- `docker/build` → `npm run build:dcr`

**Recommended shortcuts:**
- `zxb` → `build`
- `zxt` → `test`
- `zxtu` → `test/unit`
- `zxf` → `fmt`
- `zxdd` → `docs/dev`

## Subtasks
- [x] Clone repo
- [x] Document build/package config
- [x] Document CI workflows
- [x] List required tools
- [x] Assess PI feature coverage
- [x] Document gaps

## Blocked By
