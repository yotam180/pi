# PI

A developer automation platform for teams managing complex repositories. PI replaces shell shortcut files and setup scripts with a structured, polyglot, and shareable automation model.

## Install

```bash
go install github.com/vyper-tooling/pi/cmd/pi@latest
```

Or build from source:

```bash
make build       # builds bin/pi
make install     # copies to /usr/local/bin/pi
```

## Quick Start

1. Create a `.pi/` folder in your repo
2. Add automation YAML files (e.g. `.pi/docker/up.yaml`)
3. Run them: `pi run docker/up`

See `docs/README.md` for the full product definition.

## Development

```bash
make build    # compile
make test     # run tests with race detector
make clean    # remove build artifacts
```
