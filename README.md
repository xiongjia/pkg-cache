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

### Build

```bash
go build -o pkg_cache.exe ./cmd/pkgcache
```

### Run

```bash
./pkg_cache.exe
```

Or with custom config:

```bash
./pkg_cache.exe -c /path/to/config.yaml
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
cd docker
docker build -t pkg_cache .
```

### Run with Docker Compose

```bash
cd docker
docker compose up -d
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
