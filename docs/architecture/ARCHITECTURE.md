# Architecture

## Hexagonal Architecture Overview

This project is built following **Hexagonal Architecture** (Ports & Adapters) principles with elements of **Domain-Driven Design**.

## Layers

```mermaid
graph TD
    subgraph External
        HTTP[HTTP Clients]
        DB[(SQLite)]
        Redis[(Redis)]
    end

    subgraph Ports
        HTTPPort[HTTP Handlers]
    end

    subgraph Application
        Commands[Command Handlers]
        Queries[Query Handlers]
    end

    subgraph Domain
        Aggregates[Aggregates]
        ValueObjects[Value Objects]
        DomainEvents[Domain Events]
        RepoInterfaces[Repository Interfaces]
    end

    subgraph Infrastructure
        Persistence[Persistence Adapters]
        TokenService[Token Service]
        SessionStore[Session Store]
        EventBus[Event Bus]
    end

    HTTP --> HTTPPort
    HTTPPort --> Commands
    HTTPPort --> Queries
    Commands --> Aggregates
    Queries --> RepoInterfaces
    Aggregates --> DomainEvents
    Aggregates --> ValueObjects
    Commands --> RepoInterfaces
    RepoInterfaces -. implemented by .-> Persistence
    Persistence --> DB
    HTTPPort -. session .-> SessionStore
    SessionStore --> Redis
    EventBus --> Redis
    Commands --> EventBus
```

## Dependency Rule

Dependencies go **inward**:
- **Domain** — doesn't depend on anything external
- **Application** — depends only on Domain
- **Infrastructure** — depends on Domain and Application (implements interfaces)
- **Ports** — depend on Application (HTTP handlers call use cases)

## Conventions

### Bounded Contexts
- Each context is a separate package in `internal/`
- **No cross-context imports** between bounded contexts
- Communication between contexts — only through `eventbus.Bus`

### Naming
- Domain: `user.go`, `role.go`, `events.go`, `errors.go`
- Application: `command/register_user.go`, `query/get_user.go`
- Infrastructure: `persistence/user_repository.go`, `token/jwt_service.go`
- Ports: `http/handler.go`, `http/middleware.go`

### Package Structure

- `internal/{context}/`
  - `domain/` - Aggregates, Value Objects, Events, Interfaces
  - `application/` - Command/Query Handlers
    - `command/`
    - `query/`
  - `infrastructure/` - Implementations (DB, external services)
    - `persistence/`
    - `token/`
    - `session/`
  - `ports/` - Entry points (HTTP, gRPC, CLI)
    - `http/`