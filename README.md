# PI

A developer automation platform for teams managing complex repositories. PI replaces shell shortcut files and setup scripts with a structured, polyglot, and shareable automation model.

## Install

### Homebrew (macOS / Linux)

```bash
brew install yotam180/pi/pi
```

### Go

```bash
go install github.com/vyper-tooling/pi/cmd/pi@latest
```

### Pre-built binaries

Download from [GitHub Releases](https://github.com/yotam180/pi/releases).

### From source

```bash
pi run build      # builds bin/pi
pi run install    # copies to /usr/local/bin/pi
```

Or without PI installed:

```bash
go build -o bin/pi ./cmd/pi
```

## Quick Start

1. Create a `.pi/` folder in your repo
2. Add automation YAML files (e.g. `.pi/docker/up.yaml`)
3. Run them: `pi run docker/up`

See `docs/README.md` for the full product definition.

## Development

```bash
pi setup          # install dev dependencies (tsx) + shell shortcuts
pi run build      # compile (dev build with version tag)
pi run test       # run tests with race detector
pi run lint       # run go vet
pi run check      # lint + test
pi run snapshot   # local GoReleaser snapshot build (all platforms)
pi run clean      # remove build artifacts
```

PI dogfoods itself — all development workflows are defined in `.pi/` and `pi.yaml`.
