# Arch Linux

Install cctg via PKGBUILD.

## Build and Install

```bash
makepkg -si
```

## Post-Install

Config is installed at `/etc/cctg/config.yaml`. Copy and edit for your user:

```bash
mkdir -p ~/.config/cctg
cp /etc/cctg/config.yaml ~/.config/cctg/
```

Create `~/.config/cctg/.env`:
```bash
echo "TELEGRAM_BOT_TOKEN=your-token" > ~/.config/cctg/.env
chmod 600 ~/.config/cctg/.env
```

## Enable Service

```bash
systemctl --user enable --now cctg
```

## AUR

After publishing to AUR:
```bash
yay -S cctg
```
