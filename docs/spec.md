# pkg_cache Specification

## Overview

- **Project Name**: pkg_cache
- **Type**: HTTP Reverse Proxy with Caching
- **Core Functionality**: Cache HTTP/git package downloads (especially `.tgz` files from GitHub) to reduce repeated downloads
- **Target Users**: Developers using Verdaccio for npm package caching who also need to cache git-based dependencies

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

## Functionality Specification

### Core Features

1. **Configurable Caching**
   - Define URL patterns to cache via YAML config file
   - Default patterns for http/git dependencies:
     - `github\.com/.*\.tgz$` - GitHub tarball downloads
     - `github\.com/.*/archive/.*\.tar\.gz$` - GitHub archive downloads
   - Pass-through all other requests (no caching)

2. **Cache Storage**
   - Store cached files on disk at `/cache` directory
   - Use URL SHA256 hash as filename for cache lookup
   - Support cache size limit (default: 10GB)
   - LRU eviction when cache is full

3. **Parent Proxy Support**
   - Support HTTP/HTTPS parent proxy
   - Configure via environment variables
   - Support authentication (basic auth)

4. **Docker Compose Integration**
   - Run as sidecar container alongside Verdaccio
   - Listen on port 8080
   - Both services are peer-level, each handling different cache needs

### Configuration Files

**config.yaml** (default):
```yaml
cache_rules:
  - pattern: "github\\.com/.*\\.tgz$"
    enabled: true
  - pattern: "github\\.com/.*/archive/.*\\.tar\\.gz$"
    enabled: true

cache:
  directory: /cache
  max_size_mb: 10240
```

**Environment Variables**:

| Variable | Description | Default |
|----------|-------------|---------|
| PORT | Proxy listen port | 8080 |
| CACHE_DIR | Cache directory | /cache |
| CACHE_SIZE_MB | Max cache size (MB) | 10240 |
| PARENT_PROXY | Parent proxy URL | (empty) |
| PARENT_PROXY_AUTH | Parent proxy auth (base64) | (empty) |
| CONFIG_PATH | Config file path | /app/config.yaml |
| BYPASS_CACHE | Bypass cache (dev mode) | false |

### API Endpoints

- `GET /health` - Health check, returns 200 OK
- `GET /cache/stats` - Cache statistics (size, count, hit rate)

### Request Flow

1. Client requests `https://github.com/user/repo/archive/v1.0.0.tgz`
2. Proxy checks if URL matches any cache rule pattern
3. If matches: check if file exists in cache
4. If cached: serve from disk immediately
5. If not cached:
   - Download from origin (via parent proxy if configured)
   - Save to cache
   - Return to client
6. If no rule matches: proxy request directly without caching

## Docker Configuration

### Dockerfile

Multi-stage build:
- Build stage: golang:1.21-alpine
- Runtime stage: alpine:3.19
- Run as non-root user (nobody)

### Docker Compose

```yaml
services:
  verdaccio:
    image: verdaccio/verdaccio
    ports:
      - "4873:4873"
    volumes:
      - verdaccio:/verdaccio

  proxy:
    build: ./docker
    ports:
      - "8080:8080"
    volumes:
      - cache:/cache
    environment:
      - PARENT_PROXY=http://host.docker.internal:7890

volumes:
  verdaccio:
  cache:
```

## Acceptance Criteria

1. Proxy starts successfully on port 8080
2. Health endpoint returns 200 OK
3. First request downloads from GitHub and caches locally
4. Second request serves from cache (faster response)
5. Non-matching requests pass through without caching
6. Docker Compose starts both Verdaccio and proxy
7. Parent proxy configuration works
8. Cache persists across restarts
9. Cache size limit is enforced with LRU eviction
