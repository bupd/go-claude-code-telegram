# Plan: go-claude-code-telegram (cctg)

## Overview
A Go Telegram bot that bridges Claude Code CLI with Telegram users. Claude Code calls this CLI to send messages and wait for user replies.

## Design Decisions (Confirmed)
- **IPC**: Unix socket at `~/.config/cctg/cctg.sock`
- **Config**: `~/.config/cctg/config.yaml`
- **Env file**: `~/.config/cctg/.env`
- **Directory structure**: Go conventional with `internal/`
- **Reply matching**: Reply-to message if available, FIFO fallback
- **Auth**: Only allowed user IDs from config can reply
- **Formatting**: Plain text only
- **Long messages**: Fail with error (4096 char limit)
- **Notifications**: Send status messages (daemon start/stop, timeout)
- **Binary name**: `cctg`
- **Container**: Scratch base image, rootless Podman
- **Systemd**: Podman quadlet for user-level service

## Commands
```bash
cctg serve              # Run Telegram bot daemon
cctg send               # Send message (stdin), wait for reply (stdout)
cctg status             # Check if daemon is running
cctg list               # List configured sessions
```

## Directory Structure
```
go-claude-code-telegram/
├── cmd/
│   └── cctg/
│       ├── main.go
│       └── cmd/
│           ├── root.go      # Cobra root + config init
│           ├── serve.go     # serve subcommand
│           ├── send.go      # send subcommand
│           ├── status.go    # status subcommand
│           └── list.go      # list subcommand
├── internal/
│   ├── config/
│   │   └── config.go        # Config types + Viper loading
│   ├── telegram/
│   │   └── bot.go           # Telegram bot handler
│   ├── session/
│   │   └── manager.go       # Session lookup + pending queue
│   └── ipc/
│       ├── server.go        # Unix socket server
│       └── client.go        # Unix socket client
├── config.yaml.example
├── .env.example
├── go.mod
├── go.sum
├── CLAUDE.md
├── LICENSE
├── Dockerfile
├── compose.yaml
├── deploy/
│   ├── systemd/
│   │   └── cctg.service         # Direct systemd user service
│   └── quadlet/
│       └── cctg.container       # Podman quadlet for systemd
└── pkg/
    └── arch/
        └── PKGBUILD             # Arch Linux package
```

## Config Format
```yaml
telegram:
  bot_token: "your-bot-token"
  allowed_users:
    - 123456789
    - 987654321

timeout: 300  # seconds (default 5 min)

sessions:
  - name: "api"
    chat_id: -100111111
    working_dir: "/home/user/projects/api"

  - name: "frontend"
    chat_id: -100222222
    working_dir: "/home/user/projects/frontend"
```

## Implementation Steps

### Step 1: Project initialization
- Create go.mod with module path and dependencies
- Set up directory structure
- Create main.go entry point

### Step 2: Config package (`internal/config/config.go`)
- Define Config, TelegramConfig, SessionConfig structs
- Implement Viper-based loading
- Support config file search order: --config flag, ./config.yaml, ~/.config/cctg/config.yaml

### Step 3: Session manager (`internal/session/manager.go`)
- Session lookup by name and working directory
- PendingMessage struct with response channel
- Thread-safe queue per session
- Reply matching: check reply-to first, then FIFO

### Step 4: IPC layer (`internal/ipc/`)
- server.go: Unix socket server for serve command
- client.go: Unix socket client for send command
- JSON protocol for request/response

### Step 5: Telegram bot (`internal/telegram/bot.go`)
- Initialize with bot token
- Validate sender against allowed_users
- Send status notifications (start/stop/timeout)
- Route replies to pending messages
- Reject messages over 4096 chars with error

### Step 6: Cobra commands (`cmd/cctg/cmd/`)
- root.go: Global flags (--config, --session, --timeout), Viper init
- serve.go: Start daemon with graceful shutdown
- send.go: Read stdin, connect to socket, print reply/timeout
- status.go: Check socket connection
- list.go: Print configured sessions

### Step 7: Graceful shutdown and edge cases
- Signal handling (SIGINT, SIGTERM)
- Clean socket file on exit
- Handle daemon not running (return timeout message silently)

### Step 8: Deployment files
- Dockerfile with multi-stage build (scratch base)
- compose.yaml for Docker/Podman Compose
- Podman quadlet for rootless systemd integration

## Key Types

```go
// config.go
type Config struct {
    Telegram TelegramConfig  `mapstructure:"telegram"`
    Timeout  int             `mapstructure:"timeout"` // seconds
    Sessions []SessionConfig `mapstructure:"sessions"`
}

type TelegramConfig struct {
    BotToken     string  `mapstructure:"bot_token"`
    AllowedUsers []int64 `mapstructure:"allowed_users"`
}

type SessionConfig struct {
    Name       string `mapstructure:"name"`
    ChatID     int64  `mapstructure:"chat_id"`
    WorkingDir string `mapstructure:"working_dir"`
}
```

```go
// session/manager.go
type PendingMessage struct {
    ID         string
    TgMsgID    int  // Telegram message ID for reply-to matching
    Content    string
    ResponseCh chan string
    CreatedAt  time.Time
}
```

```go
// ipc/protocol.go
type Request struct {
    Type    string `json:"type"`    // "send"
    Session string `json:"session"` // session name or empty for auto-detect
    Message string `json:"message"`
    Timeout int    `json:"timeout"` // seconds
    WorkDir string `json:"workdir"` // for auto-detection
}

type Response struct {
    Success bool   `json:"success"`
    Reply   string `json:"reply"`
    Error   string `json:"error,omitempty"`
}
```

## Usage Examples
```bash
# Start daemon
cctg serve

# Send message (auto-detect session by cwd)
echo "Should I refactor this?" | cctg send

# Send with explicit session
echo "Deploy?" | cctg send --session api

# Send with custom timeout
echo "Review PR?" | cctg send --timeout 600

# Check daemon status
cctg status

# List sessions
cctg list
```

## Default Timeout Message
```
user didn't reply go ahead with caution, don't make huge refactor, check what you are doing
```

## Dependencies
```
github.com/go-telegram-bot-api/telegram-bot-api/v5
github.com/spf13/cobra
github.com/spf13/viper
github.com/joho/godotenv
```

## Environment Variables (.env)
```bash
# ~/.config/cctg/.env
TELEGRAM_BOT_TOKEN=your-bot-token
```
Viper will read from .env file and allow env vars to override config.yaml values.

## Deployment

### Dockerfile (multi-stage, scratch)
```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o cctg ./cmd/cctg

FROM scratch
COPY --from=builder /app/cctg /cctg
ENTRYPOINT ["/cctg"]
CMD ["serve"]
```

### compose.yaml
```yaml
services:
  cctg:
    build: .
    restart: unless-stopped
    env_file:
      - ~/.config/cctg/.env
    volumes:
      - ~/.config/cctg:/config:ro
      - cctg-socket:/run/cctg
    command: ["serve", "--config", "/config/config.yaml"]

volumes:
  cctg-socket:
```

### Systemd User Service (direct, no container)
```ini
# deploy/systemd/cctg.service
# When installed via PKGBUILD: /usr/lib/systemd/user/cctg.service
# For manual install: ~/.config/systemd/user/cctg.service
[Unit]
Description=Claude Code Telegram Bot
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/usr/bin/cctg serve
Restart=on-failure
RestartSec=5
EnvironmentFile=%h/.config/cctg/.env

[Install]
WantedBy=default.target
```

For manual installation (binary in ~/.local/bin), create a copy with:
```ini
ExecStart=%h/.local/bin/cctg serve
```

### Enable Direct Systemd Service
```bash
# Copy service file
mkdir -p ~/.config/systemd/user
cp deploy/systemd/cctg.service ~/.config/systemd/user/

# Enable and start
systemctl --user daemon-reload
systemctl --user enable --now cctg
systemctl --user status cctg
journalctl --user -u cctg -f
```

### Podman Quadlet (~/.config/containers/systemd/cctg.container)
```ini
[Unit]
Description=Claude Code Telegram Bot
After=network-online.target

[Container]
Image=localhost/cctg:latest
EnvironmentFile=%h/.config/cctg/.env
Volume=%h/.config/cctg:/config:ro
Volume=cctg-socket.volume:/run/cctg
Exec=serve --config /config/config.yaml

[Service]
Restart=always

[Install]
WantedBy=default.target
```

### Quadlet Volume (~/.config/containers/systemd/cctg-socket.volume)
```ini
[Volume]
```

### Enable Service
```bash
# Generate and enable
systemctl --user daemon-reload
systemctl --user enable --now cctg
systemctl --user status cctg
```

### PKGBUILD (Arch Linux)
```bash
# pkg/arch/PKGBUILD
pkgname=cctg
pkgver=0.1.0
pkgrel=1
pkgdesc='Claude Code Telegram Bot - bridge between Claude Code CLI and Telegram'
arch=('x86_64')
url="https://github.com/bupd/go-claude-code-telegram"
license=('MIT')
makedepends=('go')
backup=('etc/cctg/config.yaml')
source=("$pkgname-$pkgver.tar.gz::$url/archive/v$pkgver.tar.gz")
sha256sums=('SKIP')

prepare() {
  cd "$pkgname-$pkgver"
  mkdir -p build/
}

build() {
  cd "$pkgname-$pkgver"
  export CGO_CPPFLAGS="${CPPFLAGS}"
  export CGO_CFLAGS="${CFLAGS}"
  export CGO_CXXFLAGS="${CXXFLAGS}"
  export CGO_LDFLAGS="${LDFLAGS}"
  export GOFLAGS="-buildmode=pie -trimpath -ldflags=-linkmode=external -mod=readonly -modcacherw"
  go build -o build/cctg ./cmd/cctg
}

check() {
  cd "$pkgname-$pkgver"
  go test ./...
}

package() {
  cd "$pkgname-$pkgver"
  # Binary
  install -Dm755 build/cctg "$pkgdir/usr/bin/cctg"

  # Config (backup=() ensures it won't be overwritten on upgrade)
  install -Dm644 config.yaml.example "$pkgdir/etc/cctg/config.yaml"

  # Systemd user service (user runs: systemctl --user enable cctg)
  install -Dm644 deploy/systemd/cctg.service "$pkgdir/usr/lib/systemd/user/cctg.service"

  # License
  install -Dm644 LICENSE "$pkgdir/usr/share/licenses/$pkgname/LICENSE"
}
```

Note: Per Arch Wiki, user services from packages go to `/usr/lib/systemd/user/`, not `/etc/systemd/`.

### Install from AUR/local
```bash
# Build and install locally
cd pkg/arch
makepkg -si

# Or install from AUR (after publishing)
yay -S cctg
```

## Files to Create (in order)
1. `go.mod`
2. `internal/config/config.go`
3. `internal/session/manager.go`
4. `internal/ipc/server.go`
5. `internal/ipc/client.go`
6. `internal/telegram/bot.go`
7. `cmd/cctg/cmd/root.go`
8. `cmd/cctg/cmd/serve.go`
9. `cmd/cctg/cmd/send.go`
10. `cmd/cctg/cmd/status.go`
11. `cmd/cctg/cmd/list.go`
12. `cmd/cctg/main.go`
13. `config.yaml.example`
14. `.env.example`
15. `Dockerfile`
16. `compose.yaml`
17. `deploy/systemd/cctg.service`
18. `deploy/quadlet/cctg.container`
19. `deploy/quadlet/cctg-socket.volume`
20. `pkg/arch/PKGBUILD`
21. Update `CLAUDE.md`

## Verification

### Local testing
1. Build: `go build -o bin/cctg ./cmd/cctg`
2. Create `~/.config/cctg/config.yaml` with bot token and test chat
3. Create `~/.config/cctg/.env` with `TELEGRAM_BOT_TOKEN=xxx`
4. Run `./bin/cctg serve` in one terminal
5. Run `./bin/cctg status` to verify daemon is running
6. Run `./bin/cctg list` to see configured sessions
7. Run `echo "test" | ./bin/cctg send --session <name>` in another terminal
8. Reply in Telegram, verify response is printed to stdout
9. Test timeout by not replying within timeout period

### Container testing
1. Build image: `podman build -t cctg:latest .`
2. Run with compose: `podman-compose up -d`
3. Check logs: `podman logs cctg`

### Systemd testing (direct)
1. Install binary to `~/.local/bin/cctg`
2. Copy `deploy/systemd/cctg.service` to `~/.config/systemd/user/`
3. Run `systemctl --user daemon-reload`
4. Run `systemctl --user start cctg`
5. Check status: `systemctl --user status cctg`

### Systemd testing (Podman quadlet)
1. Copy quadlet files to `~/.config/containers/systemd/`
2. Run `systemctl --user daemon-reload`
3. Run `systemctl --user start cctg`
4. Check status: `systemctl --user status cctg`

### Arch Linux testing
1. `cd pkg/arch && makepkg -si`
2. Verify installed: `which cctg`
3. Check config backup: `pacman -Ql cctg`
