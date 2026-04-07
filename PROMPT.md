# Go DDD Hexagonal Skeleton — AI Generation Requirements

> **Призначення:** Повна специфікація для AI-генерації production-ready Go skeleton проекту.  
> **Версія:** 1.0.0

---

## 1. Мета та контекст

Згенерувати повний skeleton Go-проекту з чистою DDD/Hexagonal архітектурою. Результат — робочий, компільований, повністю задокументований і покритий тестами проект, готовий до розширення новими bounded contexts.

**Принципи генерації:**
- Генерувати тільки те, що описано у цій специфікації
- Дотримуватись ідіоматичного Go (effective Go, Go proverbs)
- Жодних зайвих абстракцій — тільки ті, що обґрунтовані архітектурно
- Кожен файл повинен компілюватись без помилок

---

## 2. Технічний стек

| Компонент | Технологія | Версія |
|-----------|------------|--------|
| Мова | Go | 1.24+ |
| HTTP Framework | Gin | v1.10+ |
| БД драйвер | sqlx | v1.3+ |
| БД | SQLite WAL | modernc.org/sqlite |
| Swagger | swaggo/swag | v1.16+ |
| Config | godotenv + os.Getenv | — |
| Логування | log/slog (stdlib) | — |
| Тести | testing + testify | v1.9+ |
| Event bus in-memory | власна реалізація | — |
| Event bus redis | go-redis/v9 | v9+ |
| UUID | google/uuid | v1.6+ |
| Міграції | golang-migrate/migrate | v4+ |

**Заборонено:** GORM, Wire/fx/dig, Cobra/Viper, будь-які фреймворки окрім Gin.

---

## 3. Структура проекту

```
skeleton/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── aux/
│   │   ├── domain/
│   │   │   └── build_info.go
│   │   ├── application/
│   │   │   └── query/
│   │   │       └── get_build_info.go
│   │   └── ports/
│   │       └── http/
│   │           └── handler.go
│   └── identity/
│       ├── domain/
│       │   ├── user.go
│       │   ├── role.go
│       │   ├── permission.go
│       │   ├── errors.go
│       │   └── events.go
│       ├── application/
│       │   ├── command/
│       │   │   ├── register_user.go
│       │   │   ├── login_user.go
│       │   │   ├── assign_role.go
│       │   │   └── revoke_role.go
│       │   └── query/
│       │       ├── get_user.go
│       │       └── list_users.go
│       ├── infrastructure/
│       │   ├── persistence/
│       │   │   ├── user_repository.go
│       │   │   └── role_repository.go
│       │   └── token/
│       │       └── jwt_service.go
│       └── ports/
│           └── http/
│               ├── handler.go
│               ├── middleware.go
│               └── dto.go
├── pkg/
│   ├── eventbus/
│   │   ├── bus.go
│   │   ├── memory/
│   │   │   └── bus.go
│   │   └── redis/
│   │       └── bus.go
│   ├── database/
│   │   └── sqlite.go
│   ├── httpserver/
│   │   └── server.go
│   ├── middleware/
│   │   ├── request_id.go
│   │   ├── logger.go
│   │   └── recovery.go
│   └── apierror/
│       └── errors.go
├── migrations/
│   ├── 001_create_users.up.sql
│   ├── 001_create_users.down.sql
│   ├── 002_create_roles.up.sql
│   ├── 002_create_roles.down.sql
│   ├── 003_create_permissions.up.sql
│   └── 003_create_permissions.down.sql
├── docs/
│   ├── architecture/
│   │   ├── ARCHITECTURE.md
│   │   ├── BOUNDED_CONTEXTS.md
│   │   ├── EVENT_BUS.md
│   │   └── RBAC.md
│   ├── adr/
│   │   ├── ADR-001-hexagonal-architecture.md
│   │   ├── ADR-002-sqlite-wal.md
│   │   ├── ADR-003-event-bus.md
│   │   ├── ADR-004-rbac-model.md
│   │   └── ADR-005-no-orm.md
│   ├── api/
│   └── development/
│       ├── GETTING_STARTED.md
│       ├── TESTING.md
│       └── CONTRIBUTING.md
├── configs/
│   ├── .env.example
│   ├── .env.test
│   └── .env.prod.example
├── scripts/
│   ├── generate_swagger.sh
│   ├── run_migrations.sh
│   └── seed_dev.sh
├── .golangci.yml
├── Makefile
├── go.mod
└── README.md
```

---

## 4. Bounded Context: `aux`

### Призначення
Надає інформацію про поточний build: версію, commit hash, час збірки, оточення. Дані **вкомпільовані** у бінарник через `ldflags`.

### Domain

```go
// internal/aux/domain/build_info.go
package domain

import "time"

// BuildInfo — value object з інформацією про збірку. Незмінний після ініціалізації.
type BuildInfo struct {
    Version   string
    Commit    string
    BuildTime time.Time
    GoVersion string
    Env       string // dev | test | prod
}

func NewBuildInfo(version, commit, buildTime, goVersion, env string) (BuildInfo, error)
```

### Ldflags injection

```go
// cmd/api/main.go
var (
    version   = "dev"
    commit    = "none"
    buildTime = "unknown"
)
```

```makefile
build:
    go build -ldflags="-X main.version=$(VERSION) -X main.commit=$(COMMIT) \
        -X main.buildTime=$(BUILD_TIME)" -o bin/api ./cmd/api
```

### HTTP Port

```
GET /api/v1/aux/info   — публічний, без автентифікації
```

Response `200`:
```json
{
  "version": "1.0.0",
  "commit": "abc1234",
  "build_time": "2025-04-07T10:00:00Z",
  "go_version": "go1.24.1",
  "env": "prod"
}
```

---

## 5. Bounded Context: `identity`

### RBAC модель

```
User ──(has many)──► UserRole ──► Role ──(has many)──► RolePermission ──► Permission
```

Permission format: `resource:action` (наприклад `users:read`, `users:write`, `roles:manage`)

**Вбудовані ролі (seed):**

| Role | Permissions |
|------|-------------|
| `super_admin` | `*:*` (wildcard) |
| `admin` | `users:read`, `users:write`, `roles:read`, `roles:manage` |
| `viewer` | `users:read` |

### Domain Layer

```go
// internal/identity/domain/user.go

type User struct {
    id           UserID
    email        Email        // value object, validated
    passwordHash PasswordHash // bcrypt, min cost 12
    roles        []RoleID
    isActive     bool
    createdAt    time.Time
    updatedAt    time.Time
}

func (u *User) AssignRole(roleID RoleID) error
func (u *User) RevokeRole(roleID RoleID) error
func (u *User) Deactivate()
func (u *User) HasPermission(permission Permission, roles []Role) bool
func (u *User) PullEvents() []DomainEvent
```

**Value Objects:** `UserID` (UUID), `Email` (validated, lowercase), `PasswordHash` (bcrypt), `Permission` (`resource:action`, validated)

**Domain Events:**
```go
type UserRegistered struct {
    UserID    UserID
    Email     Email
    OccurredAt time.Time
}

type RoleAssigned struct {
    UserID    UserID
    RoleID    RoleID
    OccurredAt time.Time
}

type RoleRevoked struct {
    UserID    UserID
    RoleID    RoleID
    OccurredAt time.Time
}
```

**Repository interfaces (ports):**
```go
type UserRepository interface {
    Save(ctx context.Context, user *User) error
    FindByID(ctx context.Context, id UserID) (*User, error)
    FindByEmail(ctx context.Context, email Email) (*User, error)
    FindAll(ctx context.Context, filter UserFilter) ([]*User, int, error)
    Delete(ctx context.Context, id UserID) error
}

type RoleRepository interface {
    Save(ctx context.Context, role *Role) error
    FindByID(ctx context.Context, id RoleID) (*Role, error)
    FindByName(ctx context.Context, name string) (*Role, error)
    FindAll(ctx context.Context) ([]*Role, error)
    FindByIDs(ctx context.Context, ids []RoleID) ([]*Role, error)
}
```

### Application Layer

**Commands:**
```
RegisterUser   { Email, Password }         → UserID, error
LoginUser      { Email, Password }         → TokenPair, error
RefreshToken   { RefreshToken }            → TokenPair, error
AssignRole     { UserID, RoleID }          → error  (requires: roles:manage)
RevokeRole     { UserID, RoleID }          → error  (requires: roles:manage)
DeactivateUser { UserID }                  → error  (requires: users:write)
```

**Queries:**
```
GetUser      { UserID }                    → UserDTO, error  (requires: users:read)
ListUsers    { Page, PageSize, Filter }    → []UserDTO, Total, error (requires: users:read)
GetMyProfile { UserID from token }         → UserDTO, error  (authenticated)
```

**Command handler структура (конвенція для всіх):**
```go
type RegisterUserHandler struct {
    users  domain.UserRepository
    roles  domain.RoleRepository
    bus    eventbus.Bus
    hasher PasswordHasher
}

func NewRegisterUserHandler(
    users domain.UserRepository,
    roles domain.RoleRepository,
    bus eventbus.Bus,
    hasher PasswordHasher,
) *RegisterUserHandler

type RegisterUserCommand struct {
    Email    string
    Password string
}

type RegisterUserResult struct {
    UserID string
}

func (h *RegisterUserHandler) Handle(ctx context.Context, cmd RegisterUserCommand) (RegisterUserResult, error)
```

### Infrastructure Layer

**JWT Token Service:**
- Access token: 15 хвилин, містить `user_id`, `roles`, `permissions`
- Refresh token: UUID, 7 днів, зберігається у БД
- Algorithm: RS256 (private/public key pair)
- Keys: завантажуються з env або файлів

**SQLite persistence:**
- Всі запити через `sqlx`, named parameters
- Жодного string concatenation у SQL
- Транзакції для операцій що змінюють кілька таблиць

### HTTP Endpoints

```
POST   /api/v1/auth/register          — публічний
POST   /api/v1/auth/login             — публічний
POST   /api/v1/auth/refresh           — публічний
POST   /api/v1/auth/logout            — authenticated

GET    /api/v1/users                  — permission: users:read
GET    /api/v1/users/:id              — permission: users:read
GET    /api/v1/users/me               — authenticated
PATCH  /api/v1/users/:id/deactivate   — permission: users:write

GET    /api/v1/roles                  — permission: roles:read
POST   /api/v1/roles                  — permission: roles:manage
POST   /api/v1/users/:id/roles        — permission: roles:manage
DELETE /api/v1/users/:id/roles/:rid   — permission: roles:manage
```

**RBAC Middleware usage:**
```go
r.GET("/users",
    authMiddleware.Authenticate(),
    rbacMiddleware.Require("users:read"),
    handler.ListUsers,
)
```

Wildcard rules: `*:*` → все дозволено. `users:*` → всі дії на `users`.

---

## 6. Event Bus

### Interface

```go
// pkg/eventbus/bus.go
package eventbus

import (
    "context"
    "time"
)

type Event interface {
    EventName() string    // "identity.user_registered"
    OccurredAt() time.Time
}

type Handler func(ctx context.Context, event Event) error

type Bus interface {
    Publish(ctx context.Context, event Event) error
    Subscribe(eventName string, handler Handler)
}
```

### In-Memory реалізація
- Синхронна доставка
- Множинні хендлери на одну подію
- Panic recovery у кожному хендлері → log error, не падати
- Використовується для `env=dev` та `env=test`

### Redis реалізація
- Pub/Sub через go-redis/v9
- Серіалізація: JSON envelope:
```json
{
  "event_name": "identity.user_registered",
  "occurred_at": "2025-04-07T10:00:00Z",
  "payload": { ... }
}
```
- Graceful shutdown: відписка від каналів
- Використовується для `env=prod`

### Вибір реалізації у wiring point

```go
// cmd/api/main.go
var bus eventbus.Bus
switch cfg.Env {
case "prod":
    bus = redisbus.New(redisClient)
default:
    bus = membus.New()
}
```

### Правила між контекстами
- Контекст A **ніколи** не імпортує пакети контексту B
- Обмін — виключно через `eventbus.Bus`
- Хендлери реєструються у `main.go`
- Назви подій: `{context}.{event_name}` (snake_case)

---

## 7. Конфігурація

### Env файли
```
configs/.env.example        — шаблон для розробки
configs/.env.test           — тестове оточення (SQLite :memory:)
configs/.env.prod.example   — шаблон для production
```

### Змінні
```bash
# Server
APP_ENV=dev          # dev | test | prod
APP_PORT=8080
APP_NAME=skeleton

# Database
DB_PATH=./data/skeleton.db   # для test: ":memory:"
DB_MAX_OPEN_CONNS=1          # SQLite: завжди 1

# Auth (RS256)
JWT_PRIVATE_KEY_PATH=./keys/private.pem
JWT_PUBLIC_KEY_PATH=./keys/public.pem
JWT_ACCESS_TTL_MINUTES=15
JWT_REFRESH_TTL_DAYS=7

# Redis (тільки prod)
REDIS_URL=redis://localhost:6379/0

# Logging
LOG_LEVEL=info       # debug | info | warn | error
LOG_FORMAT=json      # json | text
```

### Config struct
```go
// pkg/config/config.go
type Config struct {
    App      AppConfig
    Database DatabaseConfig
    Auth     AuthConfig
    Redis    RedisConfig
    Log      LogConfig
}

func Load(envFile string) (*Config, error)
```

---

## 8. HTTP Server та Middleware

### Middleware chain (порядок важливий)
1. `Recovery` — panic → 500, логування
2. `RequestID` — генерує/пробрасовує `X-Request-ID`
3. `Logger` — structured slog, включає request_id, latency, status
4. `CORS` — configurable через env
5. `Authenticate` — JWT validation (тільки захищені роути)
6. `RequirePermission` — RBAC check (після Authenticate)

### Error responses (RFC 7807)
```json
{
  "type": "https://skeleton.app/errors/validation",
  "title": "Validation Error",
  "status": 422,
  "detail": "email: must be a valid email address",
  "instance": "/api/v1/auth/register",
  "request_id": "01HZ..."
}
```

| Тип | HTTP статус |
|-----|-------------|
| `validation` | 422 |
| `unauthorized` | 401 |
| `forbidden` | 403 |
| `not_found` | 404 |
| `conflict` | 409 |
| `internal` | 500 |

---

## 9. Тести

### Пріоритет генерації

**P0 — критичні (генерувати першими):**
1. `identity/domain/user_test.go` — aggregate logic, password hash, role assignment
2. `identity/domain/permission_test.go` — wildcard matching, format validation
3. `identity/application/command/register_user_test.go` — happy path + duplicate email
4. `identity/application/command/login_user_test.go` — correct/wrong password
5. `pkg/eventbus/memory/bus_test.go` — publish/subscribe, multiple handlers, panic recovery

**P1 — важливі:**
6. `identity/infrastructure/token/jwt_service_test.go` — sign, verify, expired, tampered
7. `identity/ports/http/handler_test.go` — integration з реальним SQLite `:memory:`
8. `aux/domain/build_info_test.go`
9. `aux/ports/http/handler_test.go`

**P2 — повнота покриття:**
10. `pkg/eventbus/redis/bus_test.go` — з miniredis
11. `identity/infrastructure/persistence/user_repository_test.go`
12. `identity/infrastructure/persistence/role_repository_test.go`

### Підхід
- Unit tests: domain + application, без інфраструктурних залежностей
- Integration tests: persistence + HTTP, реальний SQLite `:memory:`
- Жодного global state між тестами
- Table-driven tests де можливо
- Pattern: Arrange → Act → Assert
- `testify/require` (не `assert`) для fatal failures

### Test helpers
```go
// internal/testutil/db.go
func NewTestDB(t *testing.T) *sqlx.DB  // SQLite :memory: + міграції

// internal/testutil/fixtures.go
func CreateTestUser(t *testing.T, repo domain.UserRepository) *domain.User
func CreateTestRole(t *testing.T, repo domain.RoleRepository, perms ...string) *domain.Role
```

### Coverage targets
- Domain layer: мінімум 90%
- Application layer: мінімум 80%
- HTTP handlers: мінімум 70%

---

## 10. SQLite WAL + Migrations

### WAL setup
```go
// pkg/database/sqlite.go
pragmas := []string{
    "PRAGMA journal_mode=WAL",
    "PRAGMA synchronous=NORMAL",
    "PRAGMA foreign_keys=ON",
    "PRAGMA busy_timeout=5000",
}
// db.SetMaxOpenConns(1)
// db.SetMaxIdleConns(1)
```

### Schema
```sql
-- 001_create_users.up.sql
CREATE TABLE users (
    id            TEXT PRIMARY KEY,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    is_active     INTEGER NOT NULL DEFAULT 1,
    created_at    TEXT NOT NULL,
    updated_at    TEXT NOT NULL
);
CREATE INDEX idx_users_email ON users(email);

-- 002_create_roles.up.sql
CREATE TABLE roles (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    created_at  TEXT NOT NULL
);

CREATE TABLE permissions (
    id       TEXT PRIMARY KEY,
    name     TEXT NOT NULL UNIQUE,  -- "users:read"
    resource TEXT NOT NULL,
    action   TEXT NOT NULL
);

CREATE TABLE role_permissions (
    role_id       TEXT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id TEXT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- 003_create_user_roles.up.sql
CREATE TABLE user_roles (
    user_id     TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id     TEXT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at TEXT NOT NULL,
    PRIMARY KEY (user_id, role_id)
);

-- 004_create_refresh_tokens.up.sql
CREATE TABLE refresh_tokens (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TEXT NOT NULL,
    created_at TEXT NOT NULL
);
CREATE INDEX idx_refresh_tokens_user ON refresh_tokens(user_id);
```

---

## 11. Makefile

```makefile
.PHONY: build run test lint swagger migrate seed clean keys

VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT     ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_TIME  = $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

build:
    go build \
      -ldflags="-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)" \
      -o bin/api ./cmd/api

run: build
    ./bin/api

test:
    go test ./... -timeout 30s

test-cover:
    go test ./... -coverprofile=coverage.out
    go tool cover -html=coverage.out -o coverage.html

test-race:
    go test -race ./...

test-p0:
    go test ./internal/identity/domain/... ./pkg/eventbus/memory/... -v

lint:
    golangci-lint run ./...

swagger:
    swag init -g cmd/api/main.go -o docs/api

migrate-up:
    go run ./scripts/migrate.go up

migrate-down:
    go run ./scripts/migrate.go down

seed:
    go run ./scripts/seed.go

keys:
    mkdir -p keys
    openssl genrsa -out keys/private.pem 2048
    openssl rsa -in keys/private.pem -pubout -out keys/public.pem

clean:
    rm -rf bin/ coverage.out coverage.html docs/api/
```

---

## 12. Документація `/docs`

### Обов'язкові файли

**`docs/architecture/ARCHITECTURE.md`** — огляд Hexagonal архітектури, Mermaid діаграма шарів (Domain → Application → Infrastructure → Ports), dependency rule, конвенції найменування.

**`docs/architecture/BOUNDED_CONTEXTS.md`** — таблиця: Context | Відповідальність | Публікує події | Підписується на. Правило no cross-context import. Checklist для додавання нового контексту.

**`docs/architecture/EVENT_BUS.md`** — коли події vs прямий виклик, envelope формат, naming conventions, гарантії доставки (in-memory vs redis), приклад реєстрації хендлера.

**`docs/architecture/RBAC.md`** — модель User→Role→Permission, формат permission, wildcard правила, як додати нову permission, middleware usage examples.

**`docs/adr/ADR-001` через `ADR-005`** — кожен у форматі:
```markdown
# ADR-XXX: Назва
## Статус: Accepted
## Контекст
## Рішення
## Наслідки
```

ADR теми: Hexagonal Architecture, SQLite WAL, Event Bus strategy, RBAC model, No ORM.

**`docs/development/GETTING_STARTED.md`** — prerequisites, clone+setup, `make keys`, `make migrate-up`, `make run`, curl examples для кожного endpoint.

**`docs/development/TESTING.md`** — як запускати, coverage targets, конвенції написання тестів.

### Swagger
- Анотації у всіх HTTP хендлерах
- `@Security BearerAuth` для захищених роутів
- Swagger UI доступний тільки у `dev` та `test` (`GET /swagger/*any`)
- Генерується у `docs/api/`

---

## 13. Правила генерації коду

```
1. Кожен публічний тип/функція/метод — має godoc коментар
2. Domain errors — sentinel errors у domain/errors.go кожного контексту
3. Context propagation — ctx перший параметр скрізь де може блокуватись
4. No naked returns
5. Errors wrapping: fmt.Errorf("operationName: %w", err)
6. Не використовувати init() функції
7. Не використовувати глобальні змінні (окрім ldflags у main)
8. Interface segregation: малі, focused інтерфейси
9. Кожен bounded context — окремий пакет без імпорту інших internal контекстів
10. HTTP handler struct — explicit constructor injection, жодного service locator
```

---

## 14. Seed дані

При запуску у `dev` оточенні (`scripts/seed.go`):
- Ролі: `super_admin`, `admin`, `viewer` з відповідними permissions
- Тестовий юзер: `admin@skeleton.local` / `Admin1234!` → роль `super_admin`
- Ідемпотентно: логувати що створено, пропускати якщо вже існує

---

## 15. Checklist після генерації

- [ ] `go build ./...` — компілюється без помилок
- [ ] `go vet ./...` — без попереджень
- [ ] `go test ./...` — всі тести проходять
- [ ] Кожен bounded context не імпортує інший напряму
- [ ] Всі HTTP endpoints мають swagger анотації
- [ ] Всі domain errors — named sentinel errors
- [ ] `migrations/` — всі up/down файли присутні
- [ ] `docs/` — всі описані файли присутні
- [ ] `.env.example` містить всі змінні
- [ ] `make build && make test` — успішно виконується

---

*Кінець специфікації. Версія 1.0.0*