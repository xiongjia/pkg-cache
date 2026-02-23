# pkg_cache

HTTP reverse proxy with caching for npm git dependencies.

## Overview

Verdaccio can only cache npm packages from the npm registry. This tool caches http/git dependencies (like GitHub tarball downloads) to reduce repeated downloads.

## Architecture

```
┌─────────────┐     ┌─────────────────┐     ┌────────────────┐
│   Client    │────▶│ Verdaccio       │     │ pkg_cache      │
│ (npm/pnpm)  │     │ (port 4873)     │     │ (port 8080)    │
│             │     │ npm packages    │     │ http/git deps  │
└─────────────┘     └────────┬────────┘     └───────┬────────┘
                            │                        │
                     ┌──────▼──────┐         ┌──────▼──────┐
                     │ npm registry│         │ Parent Proxy │
                     └─────────────┘         └──────┬───────┘
                                                   │
                                            ┌──────▼──────┐
                                            │ GitHub/      │
                                            │ Internet     │
                                            └─────────────┘
```

- **Verdaccio**: Caches npm packages from npm registry
- **pkg_cache**: Caches http/git dependencies (e.g., `git://github.com/...`, `https://github.com/...tgz`)

## Usage

### Install Go Task

First, install Go Task:

```bash
go install github.com/go-task/task/v3/cmd/task@latest
```

### Build

```bash
task build
```

### Run

```bash
task run
```

Or with custom config:

```bash
./pkg_cache.exe -c /path/to/config.yaml
```

### Other Commands

```bash
task test       # Run tests
task lint       # Run linter
task fmt        # Format code
task all        # Build, test, and lint
task docker-up  # Start Docker services
```

### Configuration

Create a `config.yaml` file:

```yaml
cache_rules:
  - pattern: "github\\.com/.*\\.tgz$"
    enabled: true
  - pattern: "github\\.com/.*/archive/.*\\.tar\\.gz$"
    enabled: true

cache:
  directory: ./cache
  max_size_mb: 10240

server:
  port: 8080
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| PORT | Listen port | 8080 |
| CACHE_DIR | Cache directory | /cache |
| CACHE_SIZE_MB | Cache size (MB) | 10240 |
| PARENT_PROXY | Parent proxy URL | (empty) |
| PARENT_PROXY_AUTH | Parent proxy auth | (empty) |
| CONFIG_PATH | Config file path | config.yaml |
| BYPASS_CACHE | Bypass cache | false |

## Docker

### Build

```bash
task docker-build
```

### Run with Docker Compose

```bash
task docker-up
```

This starts:
- Verdaccio on port 4873
- pkg_cache on port 8080

## API Endpoints

- `GET /health` - Health check
- `GET /cache/stats` - Cache statistics

## Cache Rules

URL patterns are defined in `config.yaml`. The proxy only caches URLs matching enabled patterns. All other requests pass through without caching.

Example patterns:
- `github\.com/.*\.tgz$` - GitHub tarball downloads
- `github\.com/.*/archive/.*\.tar\.gz$` - GitHub archive downloads
