# Contributing to Skeleton CRM

Thank you for your interest in contributing! This document provides guidelines and instructions for contributing.

## 🌟 Ways to Contribute

- **Bug Reports**: Submit issues for bugs you find
- **Feature Requests**: Suggest new features or improvements
- **Code Contributions**: Submit pull requests
- **Documentation**: Improve or add documentation
- **Testing**: Write tests or improve test coverage

---

## 🐛 Bug Reports

### Before Submitting

1. **Search existing issues** to avoid duplicates
2. **Test with latest version** to ensure bug still exists
3. **Gather information**:
   - Go version (`go version`)
   - Node version (`node --version`)
   - OS and version
   - Steps to reproduce

### Submitting Bug Report

Use the [bug report template](.github/ISSUE_TEMPLATE/bug_report.md):

```markdown
**Description**
Clear description of the bug

**Steps to Reproduce**
1. Start services with `make dev`
2. Create invoice POST /api/v1/invoices
3. ...

**Expected Behavior**
Invoice should be created

**Actual Behavior**
500 Internal Server Error

**Environment**
- Go: 1.25
- Node: 24.14.1
- OS: macOS 14.4
- Skeleton version: df9a4a7

**Logs**
```
time=2026-04-09 error="connection refused" ...
```
```

---

## 💡 Feature Requests

### Before Submitting

1. **Search existing issues** for similar requests
2. **Consider scope** - Is this within project goals?
3. **Think about implementation** - How would it work?

### Submitting Feature Request

Use the [feature request template](.github/ISSUE_TEMPLATE/feature_request.md):

```markdown
**Is your feature request related to a problem?**
Yes, currently cannot...

**Describe the solution you'd like**
Add support for...

**Describe alternatives you've considered**
A clear description of any alternative solutions

**Additional context**
Screenshots, mockups, or examples
```

---

## 🔧 Development Setup

### Prerequisites

- Go 1.25+
- Node.js 20+ (24+ recommended)
- PostgreSQL 16+
- Redis 7+
- Docker & Docker Compose

### Quick Setup

```bash
# Clone repository
git clone https://github.com/basilex/skeleton.git
cd skeleton

# Install dependencies
make install

# Start services
make db-up

# Run migrations
make db-migrate

# Seed data (optional)
make db-seed

# Start development
make backend  # Terminal 1
make frontend # Terminal 2
```

See [SETUP.md](docs/SETUP.md) for detailed instructions.

---

## 📝 Code Style

### Backend (Go)

#### Formatting

```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run
```

#### Conventions

- **Package names**: lowercase, single word (e.g., `invoicing`, not `invoice_service`)
- **File names**: snake_case (e.g., `invoice_repository.go`)
- **Function names**: PascalCase for exported, camelCase for private
- **Interfaces**: end with `-er` (e.g., `InvoiceRepository`)
- **Errors**: return errors as last return value, use `errors.New()` or `fmt.Errorf()`

#### DDD Structure

```
internal/invoicing/
├── domain/            # Entities, value objects, domain errors
│   ├── invoice.go
│   ├── invoice_line.go
│   └── errors.go
├── application/       # Use cases, commands, queries
│   ├── create_invoice.go
│   └── send_invoice.go
├── infrastructure/    # Repositories, external services
│   ├── invoice_repository.go
│   └── email_service.go
└── ports/             # HTTP handlers, DTOs
    ├── http/
    │   ├── handler.go
    │   ├── routes.go
    │   └── dtos.go
    └── repository.go
```

#### Example

```go
// Domain entity
package domain

type Invoice struct {
    ID          uuid.UUID
    CustomerID  uuid.UUID
    Lines       []*InvoiceLine
    Total       money.Money
    Status      InvoiceStatus
}

func NewInvoice(customerID uuid.UUID) (*Invoice, error) {
    if customerID == uuid.Nil {
        return nil, errors.New("customer ID is required")
    }
    return &Invoice{
        ID:         uuid.New(),
        CustomerID: customerID,
        Status:     Draft,
        CreatedAt:  time.Now(),
    }, nil
}

// Application service
package application

type CreateInvoiceUseCase struct {
    invoiceRepo domain.InvoiceRepository
    customerRepo domain.CustomerRepository
}

func (uc *CreateInvoiceUseCase) Execute(ctx context.Context, cmd CreateInvoiceCommand) (*Invoice, error) {
    // Validate customer exists
    customer, err := uc.customerRepo.GetByID(ctx, cmd.CustomerID)
    if err != nil {
        return nil, fmt.Errorf("get customer: %w", err)
    }
    
    // Create invoice
    invoice, err := domain.NewInvoice(customer.ID)
    if err != nil {
        return nil, fmt.Errorf("create invoice: %w", err)
    }
    
    // Save to repository
    if err := uc.invoiceRepo.Create(ctx, invoice); err != nil {
        return nil, fmt.Errorf("save invoice: %w", err)
    }
    
    return invoice, nil
}
```

### Frontend (TypeScript/React)

#### Formatting

```bash
# Format code
npx prettier --write .

# Run linter
npm run lint

# Type check
npx tsc --noEmit
```

#### Conventions

- **Component names**: PascalCase (e.g., `InvoiceList.tsx`)
- **File names**: PascalCase for components, camelCase for utilities
- **Use hooks**: Functional components with hooks
- **TypeScript**: Always use strict typing
- **API calls**: Use typed API client from `lib/api/`

#### Example

```tsx
// components/domain/InvoiceList.tsx
import { useState, useEffect } from 'react'
import { Card } from '@/components/ui/card'
import { apiClient } from '@/lib/api/client'
import type { Invoice } from '@shared/types/api'

interface InvoiceListProps {
  onInvoiceClick?: (invoice: Invoice) => void
}

export function InvoiceList({ onInvoiceClick }: InvoiceListProps) {
  const [invoices, setInvoices] = useState<Invoice[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    apiClient.get<Invoice[]>('/api/v1/invoices')
      .then(setInvoices)
      .catch(console.error)
      .finally(() => setLoading(false))
  }, [])

  if (loading) {
    return <div>Loading...</div>
  }

  return (
    <div className="space-y-4">
      {invoices.map(invoice => (
        <Card
          key={invoice.id}
          onClick={() => onInvoiceClick?.(invoice)}
        >
          <h3>{invoice.number}</h3>
          <p>Total: {invoice.total}</p>
        </Card>
      ))}
    </div>
  )
}
```

---

## 🧪 Testing

### Backend Tests

#### Unit Tests

```go
// domain/invoice_test.go
func TestNewInvoice(t *testing.T) {
    tests := []struct {
        name       string
        customerID uuid.UUID
        wantErr    bool
    }{
        {
            name:       "valid customer ID",
            customerID: uuid.New(),
            wantErr:    false,
        },
        {
            name:       "empty customer ID",
            customerID: uuid.Nil,
            wantErr:    true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            invoice, err := NewInvoice(tt.customerID)
            if tt.wantErr {
                require.Error(t, err)
                require.Nil(t, invoice)
            } else {
                require.NoError(t, err)
                require.NotNil(t, invoice)
                require.NotEmpty(t, invoice.ID)
            }
        })
    }
}
```

#### Integration Tests

```go
// tests/integration/invoice_test.go
func TestCreateInvoiceAPI(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    defer db.Close()
    
    router := setupTestRouter(db)
    
    // Create customer first
    customer := createTestCustomer(t, router)
    
    // Create invoice
    body := `{"customer_id":"` + customer.ID + `"}`
    req := httptest.NewRequest("POST", "/api/v1/invoices", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    // Assert
    assert.Equal(t, 201, w.Code)
    
    var invoice Invoice
    json.Unmarshal(w.Body.Bytes(), &invoice)
    assert.NotEmpty(t, invoice.ID)
    assert.Equal(t, customer.ID, invoice.CustomerID)
}
```

### Running Tests

```bash
# Backend unit tests
make test

# Backend integration tests
make test-integration

# Frontend tests
cd frontend && npm test

# Coverage report
make test-coverage
```

---

## 🔀 Pull Request Process

### 1. Create Branch

```bash
# Update dev branch
git checkout dev
git pull origin dev

# Create feature branch
git checkout -b feature/invoice-pdf

# Or for bugfix
git checkout -b fix/invoice-total
```

### 2. Make Changes

- Write clean, readable code
- Follow code style guidelines
- Add/update tests
- Update documentation

### 3. Commit Changes

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```bash
# Feature
git commit -m "feat(invoicing): add invoice PDF generation"

# Bug fix
git commit -m "fix(accounting): correct balance calculation"

# Documentation
git commit -m "docs(api): update invoice endpoint documentation"

# Refactoring
git commit -m "refactor(identity): extract session management"
```

### 4. Push & Create PR

```bash
# Push branch
git push origin feature/invoice-pdf

# Create PR on GitHub
# - Go to repository
# - Click "New Pull Request"
# - Select feature/invoice-pdf -> dev
# - Fill PR template
```

### PR Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix (non-breaking change)
- [ ] New feature (non-breaking change)
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing performed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] No new warnings
- [ ] Tests pass locally
```

### 5. Code Review

- Respond to review comments promptly
- Make requested changes
- Discuss alternatives if needed
- Keep PR focused (one feature/fix per PR)

### 6. Merge

- PR requires at least 1 approval
- All CI checks must pass
- Squash commits if needed
- Delete branch after merge

---

## 📖 Documentation

### When to Update Docs

- **Adding new feature** → Update ARCHITECTURE.md
- **Changing API** → Update API.md
- **Changing config** → Update SETUP.md
- **Adding workflow** → Update DEVELOPMENT.md

### Documentation Style

- Use clear, concise language
- Include code examples
- Add diagrams for complex concepts
- Keep line length < 120 chars

---

## 🏷️ Release Process

### Versioning

We use [SemVer](https://semver.org/): `MAJOR.MINOR.PATCH`

- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes

### Release Steps

1. Update version in `Makefile` and `package.json`
2. Update `CHANGELOG.md`
3. Create git tag: `git tag v2.1.0`
4. Push tag: `git push origin v2.1.0`
5. GitHub Actions builds and deploys

---

## 🤝 Code of Conduct

### Our Pledge

- Be respectful and inclusive
- Welcome different perspectives
- Focus on what's best for the community
- Show empathy towards others

### Unacceptable Behavior

- Harassment or discrimination
- Trolling or insulting comments
- Public or private harassment
- Publishing others' private information

### Reporting

Report issues to: conduct@skeleton.local

---

## 📞 Getting Help

- **Documentation**: [docs/](docs/)
- **GitHub Issues**: For bugs and features
- **GitHub Discussions**: For questions and discussions
- **Email**: support@skeleton.local

---

**Thank you for contributing! 🎉**
