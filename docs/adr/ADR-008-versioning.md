# ADR-008: Semantic Versioning Strategy

## Status

Accepted

## Context

Потрібно надійний спосіб версіонування бінарників та відстеження версій в API:

1. **Git hash як версія** не інформативний для користувачів
2. Потрібна підтримка різних середовищ (dev, staging, prod)
3. Потрібно легко розуміти яка версія розгорнута
4. CI/CD pipeline має легко налаштовувати версії

## Decision

Використовувати **Semantic Versioning** (SemVer) з environment suffix:

### Формат

```
MAJOR.MINOR.PATCH-STAGE
```

- `MAJOR` - breaking changes
- `MINOR` - new features, backward compatible
- `PATCH` - bug fixes
- `STAGE` - environment (dev, staging, prod)

### Приклади

- `0.1.0-dev` - developmentверсія
- `0.1.0-staging` - staging середовище
- `1.0.0-prod` - production release

### Реалізація

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

### Використання в CI/CD

```yaml
# GitHub Actions / GitLab CI
- name: Build
  run: |
    VERSION_STAGE=prod make build
    
# Або для release tag
- name: Build release
  run: |
    VERSION_MAJOR=$(echo $TAG | cut -d. -f1 | sed 's/v//')
    VERSION_MINOR=$(echo $TAG | cut -d. -f2)
    VERSION_PATCH=$(echo $TAG | cut -d. -f3)
    VERSION_STAGE=prod make build
```

## Consequences

### Positive

- ✓ Зрозумілі версії для користувачів та розробників
- ✓ Підтримка SemVer стандарту
- ✓ Легко налаштувати для різних середовищ
- ✓ Git commit зберігається для референсу
- ✓ Сумісність з release automation tools

### Negative

- Потрібно вручну оновлювати версії в Makefile для release
- CI/CD pipeline має передавати змінні середовища

### Neutral

- Зберігаємо git commit як додаткову інформацію
- Build time автоматично генерується

## Alternatives Considered

1. **Git tags only** - не підтримує staging середовища
2. **Date-based versioning** (2024.01.07) - не відповідає SemVer
3. **Auto-increment** - складніше для release management

## References

- [Semantic Versioning](https://semver.org/)
- [Go ldflags](https://pkg.go.dev/cmd/link)