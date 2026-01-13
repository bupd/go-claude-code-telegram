# cctg

Claude Code Telegram Bot - bridges Claude Code CLI with Telegram for user interaction.

## Install

```bash
go install github.com/bupd/go-claude-code-telegram/cmd/cctg@latest
```

## Setup

Interactive setup:
```bash
cctg init
```

Or with flags:
```bash
cctg init --token BOT_TOKEN --user-id 123456789 --chat-id 123456789 --session-name myproject --working-dir /path/to/project
```

### Getting IDs

- **Bot token**: Create bot via [@BotFather](https://t.me/BotFather)
- **User ID**: Send message to [@userinfobot](https://t.me/userinfobot)
- **Chat ID**: For private chat use your user ID, for groups check `https://api.telegram.org/bot<TOKEN>/getUpdates`

## Usage

```bash
# Start daemon
cctg serve

# Check status
cctg status

# Send message (reads stdin, prints reply)
echo "Should I proceed?" | cctg send --session myproject

# List sessions
cctg list
```

## Alternative Installation

- [Systemd user service](deploy/systemd/)
- [Podman Quadlet](deploy/quadlet/)
- [Docker/Podman Compose](deploy/container/)
- [Arch Linux PKGBUILD](pkg/arch/)

## License

MIT
