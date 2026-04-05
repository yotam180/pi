# install-python should respect pyenv

## Type
bug

## Status
todo

## Priority
high

## Project
standalone

## Description
`pi setup add python --version 3.13` fails even when the correct Python version is installed and active via pyenv. The builtin `install-python` automation only checks the system Python (which reports 3.9.6), ignoring pyenv's shim that resolves to 3.13.11.

### Steps to reproduce
1. Have pyenv installed with Python 3.13 active:
   ```
   $ python --version
   Python 3.13.11
   $ which python
   /Users/yotam/.pyenv/versions/3.13.11/bin/python
   ```
2. Run `pi setup add python --version 3.13`

### Expected
```
  ✓  install-python              already installed (3.13.11)
```

### Actual
```
3.9.6 does not satisfy 3.13
  →  install-python            installing...
Warning: python@3.13 3.13.12_1 is already installed and up-to-date.
3.9.6 does not satisfy 3.13
  ✗  install-python            failed
```

The version check picks up the macOS system Python (3.9.6) instead of the pyenv-managed one (3.13.11). It then tries to install via Homebrew (which already has it), and still reports the system version on the post-install check, so it fails permanently.

### Root cause (likely)
The version detection command (probably `python3 --version` or `/usr/bin/python3 --version`) is not resolving through pyenv's shims. The install-python automation needs to:

1. **Check pyenv first** — if `pyenv` is installed and has the requested version, use it
2. **Respect PATH ordering** — run `python3 --version` in a shell that has pyenv shims on PATH, not bypass them
3. **Support pyenv as an install strategy** — if the user has pyenv, `pyenv install 3.13` is a better install path than Homebrew

### Suggested fix
The version detection step should use whichever `python3` (or `python`) is first on PATH, which naturally respects pyenv, asdf, mise, etc. If the automation hardcodes a path like `/usr/bin/python3` or `/opt/homebrew/bin/python3`, that's the bug — it should just call `python3 --version` and let the shell resolve it.

For the install step, the priority should be:
1. If pyenv is available → `pyenv install <version>` + `pyenv local/global <version>`
2. If Homebrew is available → `brew install python@<version>`
3. Fallback to system package manager

## Acceptance Criteria
- [ ] `pi setup add python --version 3.13` succeeds when pyenv has 3.13 active
- [ ] Version check uses the PATH-resolved `python3`, not a hardcoded path
- [ ] If pyenv is installed, it is used as the preferred install strategy
- [ ] The automation still works on systems without pyenv (Homebrew fallback)

## Implementation Notes

## Subtasks
- [ ] 

## Blocked By
