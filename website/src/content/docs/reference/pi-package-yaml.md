---
title: pi-package.yaml
description: Reference for the optional pi-package.yaml file in automation packages
---

`pi-package.yaml` is an optional file at the root of a [PI package](/concepts/packages/) repository. It provides metadata about the package.

## Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `min_pi_version` | string | no | Minimum PI version required to use this package. |

## Example

```yaml
min_pi_version: "0.5.0"
```

## Behavior

### Version check

When `min_pi_version` is present, PI checks the running version against it at fetch time (during `pi add`, `pi setup`, or on-demand fetch). If the running PI version is older than `min_pi_version`, the fetch fails with a clear upgrade message.

### Dev builds

Dev builds (where the version string is `"dev"`) skip the version check entirely.

### Absent or empty

If `pi-package.yaml` is absent or contains no fields, the package works with any PI version. No file is required to publish a package.

## When to use it

Add `pi-package.yaml` when your package relies on PI features introduced in a specific version. For example, if your automations use `first:` blocks (introduced in a later version), set `min_pi_version` to that version so users get a clear error instead of a confusing parse failure.

```yaml
# Package uses first: blocks and automation-level env
min_pi_version: "0.4.0"
```

## File location

The file must be at the repository root — the same directory that contains the `.pi/` folder:

```
my-pi-package/
  .pi/
    docker/
      up.yaml
  pi-package.yaml      ← here
```
