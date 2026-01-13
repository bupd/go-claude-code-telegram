# CLAUDE.md

This file provides guidance to Claude Code when working with code in this repository.

## Project Overview

cctg (Claude Code Telegram Bot) - A Go-based Telegram bot that bridges Claude Code CLI with Telegram users. Claude Code calls this CLI to send messages and wait for user replies.

## Build and Run Commands

```bash
# Build
go build -o bin/cctg ./cmd/cctg

# Run daemon
./bin/cctg serve

# Send message (reads stdin)
echo "message" | ./bin/cctg send --session <name>

# Check status
./bin/cctg status

# List sessions
./bin/cctg list

# Test
go test ./...

# Coverage
go test -cover ./...
```

## Configuration

Config file: `~/.config/cctg/config.yaml`
Env file: `~/.config/cctg/.env`
Socket: `~/.config/cctg/cctg.sock`

## Code Style

- Follow standard Go conventions (gofmt, go vet)
- Use meaningful package names
- Keep functions small and focused
