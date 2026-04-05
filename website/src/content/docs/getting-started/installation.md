---
title: Installation
description: Install PI on macOS and Linux via Homebrew, curl, or manual download
---

## Homebrew (macOS and Linux)

The recommended way to install PI:

```bash
brew install yotam180/pi/pi
```

## Manual Download

Download the latest binary from [GitHub Releases](https://github.com/yotam180/pi/releases). Binaries are available for:

| OS | Architecture |
|----|-------------|
| macOS | arm64 (Apple Silicon), amd64 (Intel) |
| Linux | arm64, amd64 |

Extract the archive and place the `pi` binary somewhere in your `PATH`:

```bash
tar xzf pi_*.tar.gz
sudo mv pi /usr/local/bin/
```

## Verify Installation

```bash
pi version
```

You should see the installed version number.

## Shell Integration

After installing the binary, run `pi shell` inside any project that has a `pi.yaml` to install shell shortcuts:

```bash
pi shell
source ~/.zshrc  # or open a new terminal
```

This is optional but recommended — it enables shell shortcuts and makes `parent_shell: true` steps work. See [Shell Shortcuts](/concepts/shell-shortcuts/) for details.

:::note[Windows]
Windows support is limited. Bash, Python, and TypeScript steps require WSL or a compatible shell environment.
:::

## Next

[Run your first automation →](/getting-started/quick-start/)
