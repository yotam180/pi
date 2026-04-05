# pi init should create an example automation or schema reference

## Type
improvement

## Status
done

## Priority
medium

## Project
standalone

## Description
After running `pi init`, the `.pi/` directory is completely empty. There is no schema reference, no example, and no pointer to documentation. The user's only option is to run `pi validate` and iterate on error messages — which, while the error messages are decent, is a frustrating onboarding experience.

### Proposed Change
`pi init` should create `.pi/hello.yaml` with a working example.

## Acceptance Criteria
- [x] `pi init` creates at least one example file in `.pi/`
- [x] The example validates and can be run with `pi run`
- [x] "Next steps" output mentions the example file
- [x] Example demonstrates the most common step types (at minimum `bash:`)

## Implementation Notes
Implemented as part of the `scaffold-automation-command` task. See that task for full details.

- `pi init` now creates `.pi/hello.yaml` with content:
  ```yaml
  description: A sample automation — edit or delete this file
  bash: echo "Hello from PI!"
  ```
- Next steps output now shows `pi run hello` and `pi new build --bash "..."` 
- Uses single-step shorthand (the short form, per philosophy)
- 2 dedicated tests verify example creation and next steps output

## Subtasks
- [x] Create hello.yaml in initProject()
- [x] Update next steps output
- [x] Add tests

## Blocked By
