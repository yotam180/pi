# Docker Project PI Example

Demonstrates a realistic Docker Compose workflow automated with PI.

All automations simulate Docker operations (no real Docker required to run these examples).

## Automations

| Name | Description |
|------|-------------|
| `docker/up` | Start all containers |
| `docker/down` | Stop and remove all containers |
| `docker/logs` | Show recent container logs (accepts service name as arg) |
| `docker/build` | Build all container images |
| `docker/build-and-up` | Build then start (chains `docker/build` → `docker/up`) |

## Usage

```bash
# List all automations
pi list

# Start containers
pi run docker/up

# View logs (all or specific service)
pi run docker/logs
pi run docker/logs api

# Build and start
pi run docker/build-and-up

# Stop everything
pi run docker/down
```

## Shortcuts

If `pi shell` were implemented, these shortcuts would be available:

| Shortcut | Automation |
|----------|-----------|
| `dup` | `docker/up` |
| `ddown` | `docker/down` |
| `dlogs` | `docker/logs` |
| `dbuild` | `docker/build-and-up` |
