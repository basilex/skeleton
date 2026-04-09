# Skeleton CRM

> **Enterprise-grade Business Management System** built with Domain-Driven Design (DDD) and Hexagonal Architecture

[![Go Version](https://img.shields.io/badge/Go-1.25%2B-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Node Version](https://img.shields.io/badge/Node.js-20%2B-339933?style=flat&logo=node.js)](https://nodejs.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Production-ready CRM system for business automation including:
- 🧾 **Invoicing & Billing** - Invoice generation, payment tracking
- 📦 **Inventory Management** - Stock control, warehouses, reservations
- 👥 **Party Management** - Customers, suppliers, contacts
- 💰 **Accounting** - Chart of accounts, transactions, double-entry
- 📋 **Order Management** - Sales orders, purchase orders
- 📄 **Document Management** - PDF generation, signatures
- 📁 **File Management** - Upload, processing, storage
- 🔔 **Notifications** - Email, SMS, push, in-app

---

## 📖 Documentation

- **[Architecture](docs/ARCHITECTURE.md)** - System architecture and bounded contexts
- **[Setup Guide](docs/SETUP.md)** - Installation and configuration
- **[Development](docs/DEVELOPMENT.md)** - Development workflow
- **[Database](docs/DATABASE.md)** - Schema and migrations
- **[API Reference](docs/API.md)** - REST API documentation

---

## 🏃 Quick Start

### Prerequisites

- **Go** 1.25+
- **Node.js** 20+ (24+ recommended)
- **PostgreSQL** 16+
- **Redis** 7+
- **Docker** & **Docker Compose** (optional)

### Using Docker (Recommended)

```bash
# Clone repository
git clone https://github.com/basilex/skeleton.git
cd skeleton

# Start all services
make dev

# Services:
# - PostgreSQL: localhost:5432
# - Redis: localhost:6379
# - Backend API: localhost:8080
# - Frontend: localhost:3000
```

### Manual Setup

```bash
# 1. Install dependencies
make install

# 2. Start database services
make db-up

# 3. Run migrations
make db-migrate

# 4. Seed database (optional)
make db-seed

# 5. Start backend (terminal 1)
make backend

# 6. Start frontend (terminal 2)
make frontend
```

### Default Credentials

After running `make db-seed`:

- **Email**: `admin@skeleton.local`
- **Password**: `Admin1234!`
- **Role**: `super_admin`

---

## 📁 Project Structure

```
skeleton/
├── backend/                    # Go API server
│   ├── cmd/api/               # Application entry point
│   ├── internal/              # Bounded contexts (DDD)
│   │   ├── accounting/        # Chart of accounts, transactions
│   │   ├── audit/            # Audit logging
│   │   ├── catalog/          # Products & services catalog
│   │   ├── documents/       # Document generation & signatures
│   │   ├── files/           # File upload & processing
│   │   ├── identity/        # Users, roles, permissions, sessions
│   │   ├── inventory/       # Warehouses, stock, movements
│   │   ├── invoicing/       # Invoices & payments
│   │   ├── notifications/   # Multi-channel notifications
│   │   ├── ordering/        # Sales & purchase orders
│   │   └── parties/         # Customers, suppliers, partners
│   ├── pkg/                  # Shared packages
│   │   ├── config/         # Configuration management
│   │   ├── middleware/     # HTTP middleware
│   │   ├── money/          # Money value object (int64 cents)
│   │   └── uuid/           # UUID v7 generation
│   ├── scripts/             # Database scripts
│   │   ├── seed/           # Seed sample data
│   │   └── migrate/        # Migration tool
│   ├── migrations/          # SQL migrations (52 tables)
│   ├── tests/              # Integration tests
│   └── go.mod
├── frontend/                 # Next.js 16 application
│   ├── app/                # App Router pages
│   ├── components/         # React components
│   │   ├── ui/            # shadcn/ui components
│   │   └── domain/        # Business components
│   ├── lib/               # Utilities
│   │   └── api/          # API client & endpoints
│   ├── public/            # Static assets
│   └── package.json
├── shared/                  # Shared code between frontend/backend
│   └── types/             # TypeScript types
│       ├── money.ts       # Money class (matches Go)
│       └── api.ts         # API types for all contexts
├── scripts/               # Operational scripts
│   ├── docker/           # Build scripts
│   ├── deploy/           # Deploy scripts
│   └── ...
├── docs/                  # Documentation
├── .github/workflows/    # CI/CD pipelines
├── docker-compose.yml    # Docker services
├── Makefile              # Development commands
└── go.work               # Go workspace
```

---

## 🏗️ Architecture

### Domain-Driven Design

The system follows **DDD** principles with **Bounded Contexts**:

```
┌─────────────────────────────────────────────────────────────┐
│                      API Layer (Gin)                        │
│  REST endpoints → DTOs → Application Services → Domain     │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                   Bounded Contexts                          │
├─────────────────┬─────────────────┬─────────────────────────┤
│   Identity      │    Parties      │    Invoicing           │
│  - Users        │  - Customers    │  - Invoices            │
│  - Roles        │  - Suppliers    │  - Payments            │
│  - Sessions     │  - Partners     │  - Lines                │
├─────────────────┼─────────────────┼─────────────────────────┤
│   Accounting    │   Inventory     │    Ordering            │
│  - Accounts     │  - Warehouses   │  - Orders              │
│  - Transactions │  - Stock        │  - Lines               │
│  - Journal      │  - Movements    │  - Status              │
├─────────────────┼─────────────────┼─────────────────────────┤
│   Catalog       │   Documents     │    Files               │
│  - Items        │  - Templates    │  - Upload              │
│  - Categories   │  - Signatures   │  - Processing          │
└─────────────────┴─────────────────┴─────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│              Infrastructure Layer                            │
│  PostgreSQL │ Redis │ Email │ SMS │ Storage │ Events       │
└─────────────────────────────────────────────────────────────┘
```

### Key Design Patterns

- **DDD**: Bounded contexts, aggregates, domain events
- **Hexagonal Architecture**: Ports & adapters
- **CQRS**: Command Query Responsibility Segregation
- **Event Sourcing**: Domain events for cross-context communication
- **Repository Pattern**: Data access abstraction
- **Value Objects**: Money (int64 cents), Email, Phone

---

## 🛠️ Tech Stack

### Backend (Go)

- **Framework**: Gin (HTTP router)
- **Database**: PostgreSQL 16+ with pgx driver
- **Cache/Sessions**: Redis 7+
- **DI**: Wire (compile-time dependency injection)
- **Architecture**: DDD + Hexagonal + CQRS

### Frontend (Next.js)

- **Framework**: Next.js 16 (App Router)
- **Language**: TypeScript 5
- **UI**: shadcn/ui + Tailwind CSS
- **State**: React Context + SWR
- **Forms**: React Hook Form + Zod

### Infrastructure

- **Containerization**: Docker + Docker Compose
- **CI/CD**: GitHub Actions
- **Database Migrations**: Custom SQL migrations
- **Monitoring**: Health checks, structured logging

---

## 📊 Makefile Commands

```bash
make help                 # Show all commands

# Development
make dev                  # Start all services (Docker)
make backend             # Start backend server
make frontend            # Start frontend dev server

# Database
make db-up               # Start PostgreSQL + Redis
make db-migrate          # Run migrations
make db-seed            # Seed sample data
make db-shell           # PostgreSQL CLI

# Testing
make test               # Backend tests
make test-coverage      # With coverage report

# Build
make build              # Build all Docker images
make build-backend      # Build backend image
make build-frontend     # Build frontend image
```

---

## 🔌 API Endpoints

### Authentication

```http
POST   /api/v1/auth/register     # Register new user
POST   /api/v1/auth/login        # Login
POST   /api/v1/auth/refresh      # Refresh token
POST   /api/v1/auth/logout       # Logout
GET    /api/v1/auth/me           # Current user profile
```

### Business Entities

```http
# Customers
GET    /api/v1/customers         # List customers
POST   /api/v1/customers         # Create customer
GET    /api/v1/customers/:id     # Get customer
PUT    /api/v1/customers/:id     # Update customer

# Invoices
GET    /api/v1/invoices          # List invoices
POST   /api/v1/invoices          # Create invoice
GET    /api/v1/invoices/:id      # Get invoice
POST   /api/v1/invoices/:id/send     # Send invoice
POST   /api/v1/invoices/:id/payments # Record payment

# Orders
GET    /api/v1/orders            # List orders
POST   /api/v1/orders            # Create order
GET    /api/v1/orders/:id         # Get order
PATCH  /api/v1/orders/:id/status  # Update status

# Inventory
GET    /api/v1/warehouses        # List warehouses
POST   /api/v1/warehouses        # Create warehouse
GET    /api/v1/stock             # List stock
POST   /api/v1/stock/adjust      # Adjust stock

# Accounting
GET    /api/v1/accounts          # Chart of accounts
POST   /api/v1/accounts          # Create account
POST   /api/v1/transactions      # Record transaction
```

**Full API Reference**: [docs/API.md](docs/API.md)

---

## 🧪 Testing

### Backend Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run integration tests
make test-integration

# Run specific package
cd backend && go test ./internal/invoicing/...
```

### Frontend Tests

```bash
# Run tests
cd frontend && npm test

# Run e2e tests
cd frontend && npm run test:e2e
```

---

## 🚀 Deployment

### CI/CD Pipeline

Automatic deployment via **GitHub Actions**:

- **Push to `dev`** → Deploy to **staging**
- **Push to `main`** → Deploy to **production**

### Manual Deployment

```bash
# Build Docker images
./scripts/docker/build-all.sh v1.0.0

# Deploy to staging
./scripts/deploy/deploy-staging.sh

# Deploy to production
./scripts/deploy/deploy-production.sh
```

---

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'feat: add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

### Commit Convention

This project follows [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation
- `refactor:` - Code refactoring
- `test:` - Tests
- `chore:` - Maintenance

---

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

---

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/basilex/skeleton/issues)
- **Email**: support@skeleton.local
- **Docs**: [Documentation](docs/)

---

## 🎯 Roadmap

### v2.0 (Current)
- ✅ Monorepo structure (backend + frontend)
- ✅ 9 bounded contexts
- ✅ Money value object (int64 cents)
- ✅ PostgreSQL + Redis
- ✅ REST API (150+ endpoints)
- ✅ Next.js 16 frontend
- ✅ Docker + CI/CD

### v2.1 (Planned)
- ⏳ Real-time WebSocket notifications
- ⏳ File upload (S3, MinIO)
- ⏳ Email templates
- ⏳ PDF generation
- ⏳ Multi-tenant support

### v3.0 (Future)
- ⏳ GraphQL API
- ⏳ Event sourcing
- ⏳ Microservices architecture
- ⏳ Kubernetes deployment

---

**Built with ❤️ using Domain-Driven Design and Hexagonal Architecture**