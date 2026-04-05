# Add a command to scaffold new automations

## Type
feature

## Status
todo

## Priority
high

## Project
standalone

## Description
There is no way for a new user to discover the automation YAML schema without trial and error. When `pi init` creates an empty `.pi/` directory, the user has no examples, templates, or schema reference to work from.

I tried three incorrect formats before landing on the right one:
1. `steps: [{ run: cargo build }]` → error (run references automations, not shell commands)
2. `steps: [{ cmd: cargo build }]` → error (cmd is not a valid step type)
3. Only after the error message told me "must specify one of: bash, run, python, typescript, first" did I know what to use.

### Proposed Solution
Add a `pi new` (or `pi create`) command that scaffolds an automation file:

```
pi new build
# Creates .pi/build.yaml with a working template:
#   name: build
#   description: TODO
#   steps:
#     - bash: echo "hello world"
```

Advanced usage:
```
pi new setup/install-deps    # Creates .pi/setup/install-deps.yaml
pi new build --bash "cargo build"   # Pre-fills the command
```

### Alternative / Complementary
Even without a `pi new` command, `pi init` should create a commented example file in `.pi/` like `.pi/_example.yaml.sample` showing the full schema with comments. Something like:

```yaml
# Automation schema reference — delete this file when done.
# name: my-automation           # optional, inferred from filename
# description: What this does   # shown in `pi list`
#
# inputs:                       # optional
#   myvar:
#     type: string
#     default: "hello"
#     description: An input variable → $PI_IN_MYVAR
#
# requires:                     # optional
#   - python
#   - command: jq
#
# steps:
#   - bash: echo $PI_IN_MYVAR   # shell command
#   - run: other-automation     # reference another automation
#   - python: script.py         # run a python script
```

## Acceptance Criteria
- [ ] A new user can create their first automation without guessing the YAML format
- [ ] Either `pi new` exists, or `pi init` includes a reference/example file

## Implementation Notes

## Subtasks
- [ ] 

## Blocked By
