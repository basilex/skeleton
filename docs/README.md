# Skeleton API Documentation

> **Production-ready Go API with Domain-Driven Design and Hexagonal Architecture**

[![Go Version](https://img.shields.io/badge/Go-1.25%2B-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Architecture](https://img.shields.io/badge/Architecture-Hexagonal-DDD-blue)]()
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

---

## 📚 Documentation Index

### Getting Started

- **[Quick Start](QUICKSTART.md)** - Get up and running in 5 minutes
- **[Setup Guide](SETUP.md)** - Environment setup and configuration
- **[Development Guide](DEVELOPMENT.md)** - Development workflow and conventions

### Architecture & Design

- **[Architecture Overview](ARCHITECTURE.md)** - System architecture and bounded contexts
- **[Database Schema](DATABASE.md)** - Migrations, schema design, and data model
- **[ADR Index](adr/)** - Architecture Decision Records

### Testing & Quality

- **[Testing Strategy](TESTING.md)** - Testing approach and guidelines
- **[Testing Strategy Details](TESTING_STRATEGY.md)** - Comprehensive testing documentation

### Operations

- **[Deployment Guide](DEPLOYMENT_GUIDE.md)** - Production deployment instructions
- **[Makefile Reference](MAKEFILE_REFERENCE.md)** - Build and development commands
- **[Changelog](CHANGELOG.md)** - Version history and changes

---

## 🏗️ Architecture Overview

### Bounded Contexts

The system follows Domain-Driven Design with clear bounded contexts:

| Context | Purpose | Key Entities |
|---------|---------|--------------|
| **Identity** | Authentication & Authorization | User, Role, Session, Permission |
| **Accounting** | Financial Management | Account, Transaction, Journal Entry |
| **Parties** | Party Management | Customer, Supplier, Partner, Employee |
| **Contracts** | Contract Lifecycle | Contract, Terms, Renewal |
| **Invoicing** | Invoice Processing | Invoice, Payment, Credit Note |
| **Inventory** | Stock Management | Stock, Warehouse, Stock Take |
| **Catalog** | Product Catalog | Product, Variant, Price |
| **Documents** | Document Management | File, Version, Approval |
| **Notifications** | Multi-channel Alerts | Notification, Template |

### Technology Stack

- **Language**: Go 1.25+
- **Architecture**: Hexagonal (Ports & Adapters)
- **Database**: PostgreSQL 16+ with JSONB & LTREE
- **Event Bus**: In-memory (CQRS ready)
- **Containerization**: Docker & Docker Compose
- **Build**: Make + Go modules

---

## 🚀 Quick Start Guide

### Prerequisites

- Go 1.25+
- PostgreSQL 16+
- Docker & Docker Compose (optional)

### Minimal Setup

```bash
# Clone repository
git clone <repository-url>
cd skeleton

# Install dependencies
make deps

# Setup database
make db-setup

# Run migrations
make migrate

# Start server
make run
```

### Docker Setup

```bash
# Start all services
docker-compose up -d

# Run migrations
docker-compose exec api make migrate

# View logs
docker-compose logs -f api
```

### Run Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific package
go test ./internal/accounting/domain/... -v
```

---

## 📖 Core Concepts

### 1. Money Value Object

All monetary values use `pkg/money.Money` (int64 cents):

```go
// Create money values
amount, _ := money.NewFromFloat(100.50, "USD") // $100.50 USD
amount, _ := money.New(10050, "USD")            // 100.50 USD in cents

// Operations
total, _ := amount.Add(another)
diff, _ := amount.Subtract(another)

// Convert for API
floatAmount := amount.ToFloat64()  // 100.50

// Database storage (BIGINT)
// int64 in persistence layer
```

### 2. Domain Events

All aggregates publish domain events:

```go
// Create aggregate
customer, _ := domain.NewCustomer("ACME Corp", "12345678", contactInfo)

// Pull events
events := customer.PullEvents()

// Publish to event bus
for _, event := range events {
    eventBus.Publish(ctx, event)
}
```

### 3. Transaction Management

Repository operations support transaction context:

```go
// From command handler
err := txManager.Execute(ctx, func(ctx context.Context) error {
    // All operations within transaction
    if err := accountRepo.Save(ctx, account); err != nil {
        return err
    }
    return transactionRepo.Save(ctx, transaction)
})
```

---

## 🔧 Project Structure

```
skeleton/
├── cmd/
│   └── api/              # Application entry point
├── internal/
│   ├── accounting/      # Accounting bounded context
│   ├── parties/          # Parties bounded context
│   ├── invoicing/        # Invoicing bounded context
│   ├── inventory/        # Inventory bounded context
│   ├── ordering/         # Ordering bounded context
│   ├── contracts/        # Contracts bounded context
│   ├── catalog/          # Catalog bounded context
│   └── identity/         # Authentication & authorization
├── pkg/
│   ├── money/            # Money value object
│   ├── eventbus/         # Event bus interface
│   ├── transaction/      # Transaction manager
│   └── uuid/              # UUID utilities
├── migrations/           # Database migrations
├── docs/                 # Documentation
│   └── adr/              # Architecture Decision Records
└── tests/                # Integration tests
```

---

## 📊 Key Metrics

- **Lines of Code**: ~50,000+
- **Test Coverage**: 70%+ domain layer
- **Bounded Contexts**: 9
- **Domain Entities**: 50+
- **Database Tables**: 80+
- **API Endpoints**: 100+

---

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing`)
3. Make changes following [Development Guide](DEVELOPMENT.md)
4. Run tests (`make test`)
5. Commit with conventional commits
6. Push and create pull request

---

## 📝 License

This project is licensed under the MIT License - see [LICENSE](LICENSE) file.

---

## 🔗 Additional Resources

- **Main README**: [../README.md](../README.md) - Project overview
- **API Documentation**: `/swagger/index.html` (when running)
- **Architecture Records**: [adr/](adr/) - Decision documents

---

**Need help?** Check the [Development Guide](DEVELOPMENT.md) or open an issue.