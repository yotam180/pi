# Close test coverage gaps in pi info and pi validate for installer/first: block paths

## Type
improvement

## Status
done

## Priority
high

## Project
standalone

## Description

`pi info` and `pi validate` have several functions at 0% test coverage, specifically around installer automations and `first:` blocks:

**pi info (`internal/cli/info.go`)**:
- `printInstallDetail` (0%) — renders the install lifecycle summary for installer automations
- `printFirstBlockDetail` (0%) — renders `first:` block details in step details

**pi validate (`internal/cli/validate.go`)**:
- `validatePhaseFileRefs` (0%) — checks file references inside install phase steps
- `checkScalarFileRef` (0%) — checks file references in scalar install phases

These are important code paths that handle real user scenarios (installer automations are the backbone of `pi setup`). Missing test coverage means regressions could ship undetected.

## Acceptance Criteria
- [x] `printInstallDetail` has unit tests covering scalar and step-list phases, verify present/absent, version present/absent
- [x] `printFirstBlockDetail` has unit tests covering sub-step annotations, descriptions, block-level if/pipe
- [x] `validatePhaseFileRefs` has unit tests covering scalar file refs, step-list file refs, first: block file refs
- [x] `checkScalarFileRef` has unit tests covering matching and non-matching scalar values
- [x] Integration tests cover `pi info` output for installer automations
- [x] Integration test covers `pi validate` with broken installer file references
- [x] All existing tests still pass
- [x] `go build ./...` succeeds
- [x] Architecture docs updated with new test counts

## Implementation Notes

### Approach
Added 24 new tests across 4 test files to bring the four target functions from 0% to 100% coverage:

**Unit tests (internal/cli/info_test.go — 11 new tests):**
- `TestPrintInstallDetail_ScalarPhases` — scalar test/run phases, default verify, version
- `TestPrintInstallDetail_StepListPhases` — step list phases with step counts
- `TestPrintInstallDetail_WithExplicitVerify` — explicit verify phase shown instead of "(re-runs test)"
- `TestPrintInstallDetail_NoVersion` — no version line when version command empty
- `TestPrintInstallDetail_LongScalarTruncated` — truncation at 60 chars
- `TestShowAutomationInfo_InstallerType` — full integration through showAutomationInfo with installer YAML
- `TestPrintFirstBlockDetail_BasicBlock` — three sub-steps with if conditions and fallback
- `TestPrintFirstBlockDetail_WithBlockAnnotations` — block-level if: and pipe: annotations
- `TestPrintFirstBlockDetail_WithDescription` — block and sub-step descriptions
- `TestPrintFirstBlockDetail_SubStepAnnotations` — sub-step with dir, timeout, silent, env
- `TestShowAutomationInfo_FirstBlockInSteps` — full integration with first: in steps YAML

**Unit tests (internal/cli/validate_test.go — 6 new tests):**
- `TestValidate_InstallerScalarFileRef_Broken` — scalar test/run phases referencing missing .sh files
- `TestValidate_InstallerScalarFileRef_Valid` — same but with files present
- `TestValidate_InstallerStepListFileRef_Broken` — step-list phases with missing file refs
- `TestValidate_InstallerFirstBlockFileRef_Broken` — first: blocks inside install phases
- `TestValidate_InstallerVerifyPhaseFileRef` — explicit verify phase with missing file ref
- `TestValidate_InstallerInlineScriptNotFlagged` — inline scripts in install phases aren't flagged

**Integration tests (tests/integration/info_test.go — 4 new tests):**
- `TestInfo_InstallerAutomation` — pi info on install-marker shows Type: installer, lifecycle, phases
- `TestInfo_InstallerNoVersion` — pi info on install-no-version omits version line
- `TestInfo_InstallerWithInputs` — pi info on installer shows inputs section
- `TestInfo_FirstBlockSubStepDetails` — pi info shows lettered sub-steps (a., b.)

**Integration tests (tests/integration/validate_integ_test.go — 3 new tests):**
- `TestValidate_InstallerScalarFileRefBroken` — validates broken scalar file refs in installers
- `TestValidate_InstallerScalarFileRefValid` — validates valid scalar file refs pass
- `TestValidate_InstallerInlineScriptsNotFlagged` — inline installer scripts not flagged

### Coverage improvement
- CLI package: 70.6% → 76.0%
- `printInstallDetail`: 0% → 100%
- `printFirstBlockDetail`: 0% → 100%
- `validatePhaseFileRefs`: 0% → 100%
- `checkScalarFileRef`: 0% → 100%
- Total test count: 1355 → 1388

## Subtasks
- [x] Write info_test.go unit tests for printInstallDetail
- [x] Write info_test.go unit tests for printFirstBlockDetail
- [x] Write validate_test.go unit tests for validatePhaseFileRefs and checkScalarFileRef
- [x] Write integration tests for pi info on installer automations
- [x] Write integration test for pi validate with installer file refs
- [x] Run full suite, update docs

## Blocked By
