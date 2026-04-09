# Skeleton API

> **Production-ready Go API skeleton with Domain-Driven Design (DDD) and Hexagonal Architecture.**

[![Go Version](https://img.shields.io/badge/Go-1.25%2B-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Architecture](https://img.shields.io/badge/Architecture-Hexagonal-DDD-blue)]()
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

---

## 🎯 What is This?

A **battle-testedfoundation for building enterprise-grade Go applications** following Domain-Driven Design principles. This skeleton provides a complete, working CRM system that demonstrates proper DDD tactical patterns, bounded contexts, and clean architecture.

**Problem Solved:** Most Go API starters are either too simple (just CRUD) or too complex (microservices). Skeleton API hits the sweet spot - a monolithic architecture that's:

- **Domain-rich**: Business logic lives in aggregates, not anemic models
- **Testable**: 114+integration tests, domain logic isolated from infrastructure
- **Scalable**: Bounded contexts ready for extraction to microservices
- **Maintainable**: Clear separation of concerns, easy to navigate

---

## ✨ Features

### Core Business Capabilities

#### 🔐 Identity & Access Management
- **User authentication** with email/password
- **Role-Based Access Control (RBAC)** with fine-grained permissions
- **Session management** with device tracking, expiration, and revocation
- **User preferences** (theme, language, notifications, quiet hours)
- **Password security** with hashing and validation

#### 📊 Accounting & Finance
- **Chart of Accounts** with hierarchy management (parent-child relationships)
- **Journal Entries** with double-entry bookkeeping validation
- **Accounting Periods** with close/lock mechanisms
- **Bank Reconciliation** with discrepancy tracking
- **Financial reporting** ready for extension

#### 👥 Party Management
- **Customers** with credit limits and payment terms
- **Suppliers** with performance scoring and rating
- **Partners** and contact management
- **Addresses** and communication channels

#### 📋 Contract Lifecycle
- **Contract creation** with parties, terms, and dates
- **Renewal management** with auto-renewal support
- **Amendment tracking** with version history
- **Status workflow**: draft → active → expired → terminated

#### 💰 Invoicing & Payments
- **Invoice creation** with line items and tax calculation
- **Credit Notes** for invoice adjustments
- **Installment plans** with payment schedules
- **Payment tracking** and reconciliation

#### 📦 Inventory Management
- **Warehouse management** with zones and locations
- **Stock tracking** by lot and serial number
- **Stock Take** with variance detection
- **Adjustments** for gains/losses
- **Expiry tracking** for perishable goods

#### 🛍️ Product Catalog
- **Products** and **Services** with variants
- **Product Variants** (size, color, material)
- **Pricing Rules** with volume discounts
- **Category management** with hierarchy

#### 📄 Document Management
- **File uploads** with metadata
- **Version control** with change tracking
- **Approval Workflows** with multi-step processes
- **Document relationships** and linking

#### 🔔 Notifications
- **Multi-channel**: Email, SMS, Push, In-App
- **Template system** with variable substitution
- **Notification preferences** per user
- **Quiet hours** to respect user schedules
- **Delivery tracking** with retry logic

#### 💾 File Management
- **Upload/download** with streaming support
- **Virus scanning** integration
- **File type validation**
- **Processing pipeline** for transformations
- **Access control** with visibility levels

#### 📝 Audit Trail
- **Action logging** with user attribution
- **IP tracking** and user agent capture
- **Resource changes** with JSON details
- **Compliance-ready** audit trail

---

## 🏗️ Architecture

### Design Principles

**Hexagonal Architecture (Ports & Adapters)**

```
┌─────────────────────────────────────┐
│           Ports (HTTP API)          │← External world
├─────────────────────────────────────┤
│      Application Layer (CQRS-lite)  │← Use cases
├─────────────────────────────────────┤
│        Domain Layer (Aggregates)    │← Business rules
├─────────────────────────────────────┤
│   Infrastructure (DB, Cache, etc)   │← Technical details
└─────────────────────────────────────┘
```

**Domain-Driven Design Tactical Patterns:**

- **Aggregates**: Consistency boundaries (User, Invoice, Contract, etc.)
- **Value Objects**: Immutable concepts (Money, Email, Address)
- **Domain Events**: Decoupled communication between contexts
- **Repositories**: Persistence abstraction
- **Factories**: Complex object creation

### Bounded Contexts

| Context | Responsibility | Key Aggregates |
|---------|---------------|----------------|
| **Identity** | Authentication & authorization | User, Session, Role, Permission |
| **Accounting** | Financial record-keeping | Account, JournalEntry, Period |
| **Parties** | Customer/Supplier management | Customer, Supplier, Partner |
| **Contracts** | Legal agreements | Contract, Amendment |
| **Invoicing** | Billing & payments | Invoice, CreditNote, Installment |
| **Inventory** | Stock management | Stock, Lot, StockTake |
| **Catalog** | Product/Service catalog | Item, Variant, PricingRule |
| **Documents** | Document workflow | Document, ApprovalWorkflow |
| **Notifications** | Communication | Notification, Template |
| **Files** | File storage | File, Processing, Upload |
| **Audit** | Action tracking | AuditRecord |

### Tech Stack

| Layer | Technology | Purpose |
|-------|------------|---------|
| **Domain** | Pure Go | Business logic, invariants |
| **Application** | Go +CQRS-lite | Use cases, orchestration |
| **Infrastructure** | PostgreSQL 16 | Persistence, ACID transactions |
| | Redis 7 | Cache, sessions |
| **Ports** | Gin | HTTP API |
| **DI** | Wire | Dependency injection |

---

## 🚀 Quick Start

### Prerequisites

- Go1.25+
- Docker & Docker Compose
- Make (optional, but recommended)

### One-Command Setup

```bash
make fresh-start
```

This will:
1. Start PostgreSQL and Redis containers
2. Run database migrations
3. Seed initial data (admin user, roles, permissions)
4. Start the API server

**Access Points:**
- API: http://localhost:8080
- Default credentials: `admin@skeleton.local` / `Admin1234!`

### Manual Setup

```bash
# Start infrastructure
make docker-up

# Run migrations
make migrate-up

# Start development server
make dev
```

---

## 📚 Documentation

### Getting Started

| Document | Description |
|----------|-------------|
| [QUICKSTART.md](docs/QUICKSTART.md) | Quick start guide |
| [DOCKER_QUICKSTART.md](docs/DOCKER_QUICKSTART.md) | Docker setup guide |
| [DEVELOPMENT.md](docs/DEVELOPMENT.md) | Development workflow |
| [ENVIRONMENT_SETUP.md](docs/ENVIRONMENT_SETUP.md) | Environment configuration |

### Architecture & Design

| Document | Description |
|----------|-------------|
| [CROSS_CONTEXT_INTEGRATION.md](docs/CROSS_CONTEXT_INTEGRATION.md) | Context integration patterns |
| [CRM_INTEGRATION_GAPS.md](docs/CRM_INTEGRATION_GAPS.md) | Integration analysis |
| [ADR/](docs/adr/) | Architecture Decision Records |

### Operations

| Document | Description |
|----------|-------------|
| [DATABASE_MIGRATION_GUIDE.md](docs/DATABASE_MIGRATION_GUIDE.md) | Migration guide |
| [DEPLOYMENT_GUIDE.md](docs/DEPLOYMENT_GUIDE.md) | Production deployment |
| [MAKEFILE_REFERENCE.md](docs/MAKEFILE_REFERENCE.md) | Makefile command reference |

### Planning & Progress

| Document | Description |
|----------|-------------|
| [EXECUTION_PLAN.md](docs/EXECUTION_PLAN.md) | Development execution plan |
| [CONTEXTS_HARDENING.md](docs/CONTEXTS_HARDENING.md) | Context hardening progress |
| [CHANGELOG.md](docs/CHANGELOG.md) | Change history |

---

## 💻 Development

### Common Commands

```bash
# Development
make dev              # Start development server
make test             # Run all tests
make test-unit        # Run unit tests only
make test-integration # Run integration tests only
make bench            # Run benchmarks

# Code Quality
make lint             # Run golangci-lint
make fmt              # Format code

# Database
make psql             # Interactive PostgreSQL shell
make migrate-up       # Apply migrations
make migrate-down     # Rollback migrations
make db-stats         # Database statistics

# Docker
make docker-up        # Start containers
make docker-down      # Stop containers
make docker-logs      # View container logs
make fresh-start      # Clean slate setup
```

### Project Structure

```
skeleton/
├── cmd/
│   └── api/              # Application entry point
│       ├── main.go
│       ├── wire.go       # Dependency injection
│       └── routes.go     # HTTP routes
├── internal/             # Private application code
│   ├── identity/         # Bounded context: Identity
│   │   ├── domain/       # Aggregates, value objects
│   │   ├── application/  # Use cases (commands/queries)
│   │   ├── infrastructure/ # Repositories, adapters
│   │   └── ports/        # HTTP handlers
│   ├── accounting/       # Bounded context: Accounting
│   ├── parties/          # Bounded context: Parties
│   └── ...               # Other bounded contexts
├── pkg/                  # Shared utilities
│   ├── uuid/             # UUID v7 implementation
│   ├── testutil/         # Testing helpers
│   └── ...
├── migrations/           # Database migrations
├── docs/                # Documentation
└── scripts/             # Build and setup scripts
```

### Database Design

- **UUID v7** - Time-sortable primary keys for distributed systems
- **JSONB** - Queryable metadata without schema changes
- **40+ indexes** - Optimized for production workloads
- **Foreign keys** - Referential integrity
- **Check constraints** - Database-level validation

---

## 👥 Who Should Use This?

### Ideal For:

1. **Startups building SaaS products**
   - Need a solid foundation, not just CRUD
   - Wantdomain logic in the right place
   - Planning for future microservices

2. **Teams migrating from microservices to monolith**
   - Understand the pain of distributed systems
   - Want bounded context separation without operational overhead
   - Need clear module boundaries

3. **Developers learning DDD**
   - Real-world example, not toy projects
   - Proper aggregate design
   - Domain events implementation

4. **Companies building internal tools**
   - Need robust audit trails
   - Role-based access control
   - Compliance requirements

### Not Ideal For:

- Simple CRUD APIs (overkill)
- Microservices from day one (adds complexity)
- Teams unwilling to learn DDD concepts

---

## 🤝 Contributing

We welcome contributions! Here's how to get started:

### Development Workflow

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Make your changes** following our conventions:
   - Domain logic belongs in `domain/` package
   - Use cases in `application/` package
   - Infrastructure in `infrastructure/` package
   - HTTP handlers in `ports/http/`
4. **Run tests**: `make test`
5. **Run linter**: `make lint`
6. **Commit with conventional commits**:
   ```
   feat(accounting): add journal entry validation
   fix(identity): resolve session expiration issue
   docs(readme): update installation instructions
   ```
7. **Push and create a Pull Request**

### Code Conventions

- **Domain layer**: Pure Go, no external dependencies
- **Value objects**: Immutable, self-validating
- **Aggregates**: Enforce invariants, emit domain events
- **Tests**: Unit tests for domain, integration tests for repositories
- **Error handling**: Domain errors, not infrastructure exceptions

### Areas We Need Help

- 📝 **Documentation**: Tutorials, guides, examples
- 🧪 **Testing**: Improve test coverage
- 🌐 **Localization**: Multi-language support
- 🔐 **Security**: Security review and hardening
- 📊 **Performance**: Profiling and optimization
- 🔌 **Integrations**: Third-party service adapters

---

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🔗 Resources

### Learning DDD

- [Domain-Driven Design Quickly](https://www.infoq.com/minibooks/domain-driven-design-quickly/) - Free mini-book
- [DDD in Practice](https://www.youtube.com/watch?v=2VcDeH1lAgo) - Video series
- [Aggregate Design](https://www.dddcommunity.org/library/vernon_2011/) - Aggregate pattern explained

### Architecture

- [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/) - Original concept
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-components-of-clean-architecture.html) - Uncle Bob's take

### Go Best Practices

- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) - Official wiki
- [Practical Go](https://dave.cheney.net/practical-go) - Dave Cheney's guidelines

---

## 📧 Contact & Support

- **Issues**: [GitHub Issues](https://github.com/basilex/skeleton/issues)
- **Discussions**: [GitHub Discussions](https://github.com/basilex/skeleton/discussions)

---

## 🗺️ Roadmap

### Upcoming Features

- [ ] GraphQL API alongside REST
- [ ] Event sourcing option for audit context
- [ ] Multi-tenancy support
- [ ] Background job processing
- [ ] Real-time notifications via WebSockets
- [ ] API rate limiting and throttling
- [ ] OpenAPI/Swagger documentation generation

### Long-term Goals

- [ ] Admin dashboard UI
- [ ] Plugin system for extensibility
- [ ] Kubernetes deployment manifests
- [ ] Performance monitoring integration
- [ ] Message queue integration (NATS/Kafka)

---

**Built with ❤️using Go and Domain-Driven Design principles.**