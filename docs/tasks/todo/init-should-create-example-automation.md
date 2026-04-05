# pi init should create an example automation or schema reference

## Type
improvement

## Status
todo

## Priority
medium

## Project
standalone

## Description
After running `pi init`, the `.pi/` directory is completely empty. There is no schema reference, no example, and no pointer to documentation. The user's only option is to run `pi validate` and iterate on error messages — which, while the error messages are decent, is a frustrating onboarding experience.

### Current behavior
```
$ pi init --yes
Initialized project 'my-project'.

  Created pi.yaml
  Created .pi/

Next steps:
  pi setup add python --version 3.13   add a setup step
  pi shell                             install shell shortcuts
  pi run <name>                        run an automation
```

The "Next steps" section is helpful for setup and running, but there's a missing step: **creating your first automation**. A developer will naturally try to create `.pi/*.yaml` files and will have no idea what format to use.

### Proposed Change
`pi init` should create `.pi/_example.yaml` (or `.pi/hello.yaml`) with a working example:

```yaml
name: hello
description: A sample automation — delete or rename this file
steps:
  - bash: echo "Hello from PI!"
```

And the "Next steps" output should include:
```
  .pi/hello.yaml                       example automation (edit or delete)
```

This single change dramatically improves the new user experience.

## Acceptance Criteria
- [ ] `pi init` creates at least one example file in `.pi/`
- [ ] The example validates and can be run with `pi run`
- [ ] "Next steps" output mentions the example file
- [ ] Example demonstrates the most common step types (at minimum `bash:`)

## Implementation Notes

## Subtasks
- [ ] 

## Blocked By
