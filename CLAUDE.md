# Project Context for pkg-cache

## Project Overview

HTTP reverse proxy with caching for npm git dependencies. Works alongside Verdaccio to cache http/git package downloads.

## Key Files

- `cmd/pkgcache/main.go` - Entry point
- `pkg/proxy/server.go` - HTTP proxy server
- `pkg/cache/store.go` - Cache storage
- `pkg/config/config.go` - Configuration
- `config.yaml` - Default configuration

## Commands

- Build: `go build -o pkg_cache.exe ./cmd/pkgcache`
- Docker: `cd docker && docker compose up -d`

## Guidelines

- Use snake_case for file names
- Keep README.md in English
- Keep source code in English (comments, logs)
- Only commit .vscode/settings.json, not other .vscode files
