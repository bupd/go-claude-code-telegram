# Systemd User Service

Run cctg as a systemd user service.

## Prerequisites

- cctg binary installed at `/usr/bin/cctg` or `~/.local/bin/cctg`
- Config at `~/.config/cctg/config.yaml`
- Env at `~/.config/cctg/.env`

## Install

```bash
mkdir -p ~/.config/systemd/user
cp cctg.service ~/.config/systemd/user/
```

If binary is at `~/.local/bin/cctg`, edit the service:
```bash
sed -i 's|/usr/bin/cctg|%h/.local/bin/cctg|' ~/.config/systemd/user/cctg.service
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
```
