# Basic PI Example

Demonstrates core PI features with a simple project.

## Automations

| Name | Description | Features Shown |
|------|-------------|----------------|
| `greet` | Print a greeting | Inline bash step, argument passing |
| `build/compile` | Simulate compilation | `.sh` file step (script beside automation) |
| `deploy` | Build then deploy | `run:` step chaining (calls `build/compile` first) |

## Usage

```bash
# List all automations
pi list

# Run the greeting (with optional name arg)
pi run greet
pi run greet Alice

# Run the build
pi run build/compile

# Run deploy (chains: build → deploy)
pi run deploy
```
