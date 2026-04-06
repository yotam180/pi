# Typed Display Status Constants

## Type
improvement

## Status
done

## Priority
medium

## Project
standalone

## Description

The `display.Printer` methods `InstallStatus()` and `PackageFetch()` dispatch icon styling based on raw Unicode strings (`"✓"`, `"✗"`, `"→"`, `"↓"`, `"⚠"`). This has two problems:

1. **Bug:** `PackageFetch` doesn't handle the `⚠` warning icon — it falls through to plain style instead of yellow/warning style. The `⚠` icon is used for "not found" file package sources in `cli/discover.go:172`, but the user sees it unstyled. This is inconsistent with the `Warn()` method which uses yellow.

2. **Fragility:** Icon string dispatch is brittle — a typo produces no compile error, just wrong styling. As the display layer grows (more status types, potential for user-customizable themes), typed constants are needed.

**Solution:** Introduce a `StatusKind` type (string-backed enum) with named constants. Change `InstallStatus` and `PackageFetch` to accept `StatusKind` instead of raw icon strings. Each `StatusKind` maps to both an icon and a style, centralizing the rendering logic. Fix the `⚠` bug by adding `StatusWarning` with yellow styling.

**Callers to update:**
- `internal/executor/install.go` — `printInstallStatus()` (5 call sites)
- `internal/cli/discover.go` — `PackageFetch()` (6 call sites)
- `internal/cli/add.go` — `PackageFetch()` (3 call sites)
- `internal/display/display_test.go` — all test call sites

## Acceptance Criteria
- [x] `StatusKind` type with `StatusSuccess`, `StatusSuccessCached`, `StatusInProgress`, `StatusFailed`, `StatusWarning` constants
- [x] `InstallStatus` and `PackageFetch` accept `StatusKind` instead of raw icon string
- [x] `⚠` warning icon renders in yellow/warning style
- [x] All callers updated to use typed constants
- [x] Tests updated and passing
- [x] `go build ./...` and `go test ./...` pass
- [x] Architecture docs updated

## Implementation Notes

Decision: Use a string-backed enum (`type StatusKind string`) rather than iota-based int enum. This provides both type safety and human-readable values in debug output. The icon rendering is centralized in a `statusIcon()` helper.

The `StatusSuccess` vs `StatusSuccessCached` split handles the existing dimming logic for "already installed" without the caller needing to pass status text that the display layer pattern-matches on.

Callers in `executor/install.go` previously distinguished "already installed" (dim) vs "installed" (green) via the icon+status-text combination. With typed constants:
- `StatusSuccessCached` → dim (✓ already installed)
- `StatusSuccess` → bold green (✓ installed)

## Subtasks
- [x] Add `StatusKind` type and constants to `display` package
- [x] Refactor `InstallStatus` and `PackageFetch` to use `StatusKind`
- [x] Fix `⚠` warning icon styling
- [x] Update `executor/install.go` callers
- [x] Update `cli/discover.go` and `cli/add.go` callers
- [x] Update all display tests
- [x] Verify build and tests pass
- [x] Update architecture docs

## Blocked By
