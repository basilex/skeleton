# ADR-008: Semantic Versioning Strategy

## Status

Accepted

## Context

Need a reliable way to version binaries and track versions in API:

1. **Git hash as version** not informative for users
2. Need support for different environments (dev, staging, prod)
3. Need to easily understand which version is deployed
4. CI/CD pipeline should easily configure versions

## Decision

Use **Semantic Versioning** (SemVer) with environment suffix:

### Format

```
MAJOR.MINOR.PATCH-STAGE
```

- `MAJOR` - breaking changes
- `MINOR` - new features, backward compatible
- `PATCH` - bug fixes
- `STAGE` - environment (dev, staging, prod)

### Examples

- `0.1.0-dev` - development version
- `0.1.0-staging` - staging environment
- `1.0.0-prod` - production release

### Implementation

```makefile
VERSION_MAJOR ?= 0
VERSION_MINOR ?= 1
VERSION_PATCH ?= 0
VERSION_STAGE ?= dev
VERSION       = $(VERSION_MAJOR).$(VERSION_MINOR).$(VERSION_PATCH)-$(VERSION_STAGE)
COMMIT        ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")

build:
	go build \
	  -ldflags="-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)" \
	  -o bin/api ./cmd/api
```

### API Endpoint

```bash
GET /build

Response:
{
  "version": "0.1.0-dev",
  "commit": "c4410c8",
  "build_time": "2026-04-07T10:00:37Z",
  "go_version": "go1.26.1",
  "env": "dev"
}
```

### Usage in CI/CD

```yaml
# GitHub Actions / GitLab CI
- name: Build
  run: |
    VERSION_STAGE=prod make build
    
# Or for release tag
- name: Build release
  run: |
    VERSION_MAJOR=$(echo $TAG | cut -d. -f1 | sed 's/v//')
    VERSION_MINOR=$(echo $TAG | cut -d. -f2)
    VERSION_PATCH=$(echo $TAG | cut -d. -f3)
    VERSION_STAGE=prod make build
```

## Consequences

### Positive

- ✓ Clear versions for users and developers
- ✓ SemVer standard support
- ✓ Easy to configure for different environments
- ✓ Git commit preserved for reference
- ✓ Compatibility with release automation tools

### Negative

- Need to manually update versions in Makefile for release
- CI/CD pipeline must pass environment variables

### Neutral

- Keep git commit as additional information
- Build time automatically generated

## Alternatives Considered

1. **Git tags only** - doesn't support staging environments
2. **Date-based versioning** (2024.01.07) - doesn't follow SemVer
3. **Auto-increment** - more complex for release management

## References

- [Semantic Versioning](https://semver.org/)
- [Go ldflags](https://pkg.go.dev/cmd/link)