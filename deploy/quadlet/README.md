# Podman Quadlet

Run cctg as a rootless Podman container via systemd quadlet.

## Prerequisites

- Podman installed
- cctg image built: `podman build -t cctg:latest .`
- Config at `~/.config/cctg/config.yaml`
- Env at `~/.config/cctg/.env`

## Install

```bash
mkdir -p ~/.config/containers/systemd
cp cctg.container cctg-socket.volume ~/.config/containers/systemd/
```

## Enable

```bash
systemctl --user daemon-reload
systemctl --user enable --now cctg
```

## Commands

```bash
systemctl --user status cctg
systemctl --user restart cctg
journalctl --user -u cctg -f
podman logs -f systemd-cctg
```
