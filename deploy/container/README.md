# Docker/Podman Compose

Run cctg using Docker or Podman Compose.

## Prerequisites

- Docker or Podman with compose
- Config at `~/.config/cctg/config.yaml`
- Env at `~/.config/cctg/.env`

## Build and Run

```bash
# From project root
docker compose up -d
# or
podman-compose up -d
```

## Commands

```bash
docker compose logs -f
docker compose restart
docker compose down
```

## Files

- `Dockerfile` - Multi-stage build with scratch base
- `compose.yaml` - Service definition
