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
- **User ID**: Enter your `@username` during init (sends /start to bot first) or get from [@userinfobot](https://t.me/userinfobot)
- **Chat ID**:
  - Private chat: Press Enter during init to use your user ID
  - Group chat: Type `group` during init after adding bot to group and sending a message

**Note:** In group chats, Telegram's privacy settings only allow bots to see messages that are replies to the bot. To receive all messages, disable privacy via @BotFather (`/setprivacy` -> Disable), then remove and re-add the bot to the group.

## Adding More Sessions

Edit `~/.config/cctg/config.yaml` to add sessions:

```yaml
sessions:
  - name: "api"
    chat_id: 123456789
    working_dir: "/home/user/projects/api"

  - name: "frontend"
    chat_id: -100222222  # group chat
    working_dir: "/home/user/projects/frontend"
```

Each session maps a working directory to a Telegram chat. Use `--session` flag or run from the working directory for auto-detection.

## Usage

```bash
# Start daemon
cctg serve

# Check status
cctg status

# Send message (auto-detect session from cwd)
cctg send "Should I proceed?"

# Send with explicit session
cctg send --session api "Deploy?"

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
