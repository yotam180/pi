---
title: Publishing to GitHub
description: Create and publish a PI automation package on GitHub
---

This guide walks through creating a PI package repository from scratch, tagging a release, and making it available for others to use with `pi add`.

## What you'll learn

- How to structure a package repository
- How to organize automations for sharing
- How to tag and publish a release
- How consumers add your package
- How to release updates
- Best practices for composable, reusable automations

---

## Step 1 — Create the package repo

A PI package is just a GitHub repo with a `.pi/` folder — the same structure as any PI project.

```bash
mkdir my-pi-package
cd my-pi-package
git init
```

Create the `.pi/` folder:

```bash
mkdir -p .pi/docker .pi/setup
```

## Step 2 — Write your automations

Organize automations by domain or lifecycle stage:

```
.pi/
  docker/
    up.yaml
    down.yaml
    logs.yaml
  setup/
    install-deps.yaml
    configure-hooks.yaml
```

Each file is a standard PI automation:

```yaml
# .pi/docker/up.yaml
description: Start Docker containers
bash: docker compose up -d "$@"
```

```yaml
# .pi/docker/down.yaml
description: Stop Docker containers
bash: docker compose down "$@"
```

```yaml
# .pi/setup/install-deps.yaml
description: Install project dependencies
steps:
  - bash: npm ci
    if: file.exists("package-lock.json")
  - bash: pip install -r requirements.txt
    if: file.exists("requirements.txt")
```

Automation names are derived from the file path: `.pi/docker/up.yaml` becomes `docker/up`. A consumer who adds your package and writes `run: docker/up` will get your automation — unless they have a local `.pi/docker/up.yaml`, which always takes priority.

:::note
Package automations use the same YAML schema as local automations. Everything works: `steps:`, `install:`, `inputs:`, `if:`, `first:`, `env:`, `dir:`, `timeout:`, `silent:`, `pipe:`, `parent_shell:`.
:::

## Step 3 — Add `pi-package.yaml` (optional)

If your automations require a minimum version of PI, add a `pi-package.yaml` at the repo root:

```yaml
# pi-package.yaml
min_pi_version: "0.5.0"
```

When present, PI checks the running version at fetch time. If the consumer's PI is older, the fetch fails with a clear upgrade message.

If you don't need a version constraint, skip this file entirely — the package will work with any PI version.

## Step 4 — Commit and tag a release

```bash
git add .
git commit -m "initial package release"
git remote add origin git@github.com:your-org/my-pi-package.git
git push -u origin main
```

Tag the release:

```bash
git tag v1.0
git push origin v1.0
```

The version tag is what consumers reference. Without a tag, there's no stable version to pin.

:::tip
Use semantic versioning (`v1.0`, `v1.1`, `v2.0`) for release tags. Consumers pin to specific versions for reproducible builds.
:::

## Step 5 — Consumers add it

Anyone can now add your package:

```bash
pi add your-org/my-pi-package@v1.0
```

This does three things:

1. Validates the reference
2. Clones the repo at `v1.0` into `~/.pi/cache/github/your-org/my-pi-package/v1.0/`
3. Appends the entry to the consumer's `pi.yaml`:

```yaml
packages:
  - your-org/my-pi-package@v1.0
```

The consumer can now use your automations:

```bash
pi run docker/up
pi run setup/install-deps
```

Or reference them in their own automations:

```yaml
steps:
  - run: docker/up
  - run: setup/install-deps
```

## Step 6 — Releasing updates

When you make changes, commit and tag a new version:

```bash
git add .
git commit -m "add setup/configure-hooks automation"
git tag v1.1
git push origin main
git push origin v1.1
```

Consumers upgrade by changing the version in `pi.yaml`:

```yaml
packages:
  - your-org/my-pi-package@v1.1    # was @v1.0
```

Or by running `pi add` with the new version:

```bash
pi add your-org/my-pi-package@v1.1
```

## Naming conventions

**Repo names:** Use `kebab-case` and prefix with `pi-` so the package is discoverable on GitHub:

```
your-org/pi-docker-utils
your-org/pi-setup-common
your-org/pi-ci-helpers
```

**Automation paths:** Group by domain inside `.pi/`:

```
.pi/
  docker/          # container lifecycle
  setup/           # environment setup
  deploy/          # deployment workflows
  ci/              # CI-specific automations
```

## Making automations composable

Good package automations are small, single-purpose, and parameterized.

**Use `inputs:` for parameterization:**

```yaml
# .pi/deploy/push-image.yaml
description: Build and push a Docker image

inputs:
  registry:
    type: string
    description: Container registry URL
  tag:
    type: string
    description: Image tag
    default: latest

bash: |
  docker build -t "$PI_IN_REGISTRY:$PI_IN_TAG" .
  docker push "$PI_IN_REGISTRY:$PI_IN_TAG"
```

**Let callers compose with `run:`:**

Instead of building a monolithic "do everything" automation, provide focused building blocks:

```yaml
# Consumer's automation composing package automations
# .pi/deploy.yaml
description: Full deployment pipeline
steps:
  - run: ci/lint
  - run: ci/test
  - run: deploy/push-image
    with:
      registry: ghcr.io/my-org/my-app
      tag: "v2.0"
```

**Write idempotent automations:** Check before acting, especially for setup and install automations. Use `if:` to skip steps when the desired state already exists:

```yaml
# Good: idempotent
steps:
  - bash: npm ci
    if: file.exists("package-lock.json")
```

## Example: a complete package

Here's the full structure of a realistic shared package:

```
pi-team-utils/
  .pi/
    docker/
      up.yaml              # docker compose up -d
      down.yaml            # docker compose down
      logs.yaml            # docker compose logs -f
      build-and-up.yaml    # build then start
    setup/
      install-deps.yaml    # npm ci + pip install
      configure-hooks.yaml # git hooks from hooks/
    deploy/
      push-image.yaml      # build and push Docker image
  pi-package.yaml          # min_pi_version: "0.3.0"
  README.md                # usage instructions
```

Tag it:

```bash
git tag v1.0 && git push origin v1.0
```

Consumers add it:

```bash
pi add your-org/pi-team-utils@v1.0
```

## Summary

- A PI package is a GitHub repo with a `.pi/` folder — the same structure as any PI project
- Organize automations by domain: `docker/`, `setup/`, `deploy/`
- Tag releases with semantic versions for stable, reproducible consumption
- Use `pi-package.yaml` to enforce a minimum PI version if needed
- Write small, idempotent, parameterized automations — let callers compose
- Prefix repo names with `pi-` for discoverability on GitHub
