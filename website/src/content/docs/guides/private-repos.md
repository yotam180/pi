---
title: Private Repositories
description: Use PI packages from private GitHub repositories via SSH or GITHUB_TOKEN
---

This guide covers how to configure authentication so PI can fetch packages from private GitHub repositories.

## What you'll learn

- How PI authenticates when fetching GitHub packages
- How to set up SSH authentication for developer machines
- How to use `GITHUB_TOKEN` for CI environments
- How to verify your authentication setup
- How to read PI's error messages when authentication fails

---

## How PI authenticates

When PI fetches a GitHub package (via `pi add`, `pi setup`, or on-demand), it tries three authentication methods in order:

1. **SSH** — `git@github.com:org/repo.git`
2. **HTTPS with token** — uses the `GITHUB_TOKEN` environment variable
3. **Plain HTTPS** — works for public repos only

PI stops at the first method that succeeds. For public repos, all three methods work. For private repos, you need either SSH or `GITHUB_TOKEN`.

## SSH setup (recommended for developer machines)

SSH is the most common way to authenticate on a developer's machine. If you already push and pull from GitHub via SSH, PI uses the same credentials automatically.

### 1. Generate an SSH key (if you don't have one)

```bash
ssh-keygen -t ed25519 -C "your-email@example.com"
```

Press Enter to accept the default location (`~/.ssh/id_ed25519`). Optionally set a passphrase.

### 2. Add the key to GitHub

Copy the public key:

```bash
cat ~/.ssh/id_ed25519.pub
```

Go to **GitHub → Settings → SSH and GPG keys → New SSH key**. Paste the key.

### 3. Verify it works

```bash
ssh -T git@github.com
```

Expected output:

```
Hi username! You've successfully authenticated, but GitHub does not provide shell access.
```

### 4. Use PI normally

After SSH is configured, `pi add` and `pi setup` work transparently with private repos:

```bash
pi add your-org/private-package@v1.0
```

No additional configuration needed — PI tries SSH first.

## GITHUB_TOKEN setup (recommended for CI)

In CI environments where SSH keys aren't available, use the `GITHUB_TOKEN` environment variable.

### On GitHub Actions

`GITHUB_TOKEN` is automatically available in GitHub Actions. Pass it to PI:

```yaml
# .github/workflows/ci.yml
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install PI
        run: brew install yotam180/pi/pi

      - name: Setup
        run: pi setup --no-shell
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

:::note
The default `GITHUB_TOKEN` in GitHub Actions has read access to the current repository. For packages in **other** private repos, you need a personal access token (PAT) stored as a repository secret.
:::

### With a personal access token

For cross-repo access (or outside GitHub Actions):

1. Go to **GitHub → Settings → Developer settings → Personal access tokens → Fine-grained tokens**
2. Create a token with **Contents: Read** permission for the repos containing your packages
3. Set it as an environment variable:

```bash
export GITHUB_TOKEN="github_pat_your_token_here"
```

For CI, store it as a repository secret:

```yaml
- name: Setup
  run: pi setup --no-shell
  env:
    GITHUB_TOKEN: ${{ secrets.PI_PACKAGES_TOKEN }}
```

### Classic tokens

Classic personal access tokens also work. The token needs the `repo` scope for private repos.

## Verifying your setup

### Verify SSH

```bash
ssh -T git@github.com
```

If this outputs `Hi username!`, SSH is working.

### Verify GITHUB_TOKEN

```bash
curl -s -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/user | head -5
```

If this returns your GitHub user information, the token is valid.

### Test a PI fetch

Try adding a package from the private repo:

```bash
pi add your-org/private-package@v1.0
```

If authentication succeeds, the package is fetched and added to `pi.yaml`.

## Reading error messages

When all authentication methods fail, PI prints a clear error:

```
could not fetch org/repo: check network and that the repo exists.
For private repos:
  • Ensure an SSH key is configured (git@github.com:org/repo.git)
  • Or set GITHUB_TOKEN env var for HTTPS auth
```

Common causes:

| Symptom | Likely cause |
|---------|-------------|
| SSH fails with "permission denied" | SSH key not added to GitHub, or wrong key |
| HTTPS with token fails | Token expired, revoked, or lacks repo read access |
| All methods fail | Repo doesn't exist, or org name is misspelled |
| Works locally but fails in CI | `GITHUB_TOKEN` not set in the CI environment |

## Multiple private repos

If your packages come from multiple private repos across different GitHub organizations, SSH is the simplest approach — one SSH key gives access to all repos where it's been added.

With tokens, a single fine-grained PAT can be scoped to multiple repos from the same owner. For repos across organizations, you may need separate tokens — set `GITHUB_TOKEN` to the one that covers the most repos.

## Summary

- PI tries SSH first, then `GITHUB_TOKEN`, then plain HTTPS
- For developer machines, configure SSH — it works transparently
- For CI, set `GITHUB_TOKEN` in the workflow environment
- Verify SSH with `ssh -T git@github.com`
- Verify tokens with `curl -s -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/user`
- PI's error messages tell you exactly which methods were tried and what to check
