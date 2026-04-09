# Skeleton Frontend

Next.js 14 frontend for Skeleton CRM with TypeScript and Tailwind CSS.

## Prerequisites

- **Node.js** ≥ 20.x
- **npm** ≥ 10.x

## Installation

```bash
# Install dependencies
npm install

# Install Tailwind CSS and dependencies
npm install -D tailwindcss-animate

# Initialize shadcn/ui (optional, but recommended)
npx shadcn@latest init

# Add common shadcn/ui components
npx shadcn@latest add button card input table form
```

## Development

```bash
# Start development server
npm run dev

# Open http://localhost:3000
```

## Environment Variables

Create `.env.local` file:

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_APP_NAME=Skeleton CRM
```

## Project Structure

```
frontend/
├── app/                    # Next.js App Router
│   ├── page.tsx           # Home page
│   ├── layout.tsx         # Root layout
│   └── globals.css        # Global styles
├── components/
│   ├── ui/                # shadcn/ui components
│   └── domain/            # Domain-specific components
├── lib/
│   ├── api/               # API client
│   │   ├── client.ts      # Base API client
│   │   └── endpoints.ts   # API endpoints
│   └── utils/             # Utility functions
├── public/                 # Static assets
├── shared/                 # Shared TypeScript types
│   └── types/
│       ├── money.ts       # Money value object
│       └── api.ts         # API types
├── styles/
├── types/
├── package.json
├── next.config.js
├── tailwind.config.ts
└── tsconfig.json
```

## TypeScript Types

Shared types are located in `shared/types/`:

- **money.ts** - TypeScript Money class matching Go implementation
- **api.ts** - API request/response types for all bounded contexts

## API Client

The API client (`lib/api/client.ts`) provides:

- `APIClient` - Base HTTP client with auth token support
- Automatic JSON serialization
- Error handling with `APIError`
- Convenience methods: `get()`, `post()`, `put()`, `patch()`, `delete()`

API endpoints (`lib/api/endpoints.ts`) provide typed functions for:

- Authentication (`authAPI`)
- Parties/Customers (`partiesAPI`)
- Accounting (`accountingAPI`)
- Invoicing (`invoicingAPI`)
- Ordering (`orderingAPI`)
- Inventory (`inventoryAPI`)
- Catalog (`catalogAPI`)

## Testing

```bash
# Run tests
npm test

# Run tests in watch mode
npm test -- --watch
```

## Build

```bash
# Build for production
npm run build

# Start production server
npm run start
```

## Useful Scripts

```bash
# Lint code
npm run lint

# Type check
npm run type-check

# Format code
npm run format
```

## Docker

Build and run with Docker:

```bash
# Build image
docker build -t skeleton-frontend .

# Run container
docker run -p 3000:3000 skeleton-frontend
```

Or use docker-compose from project root:

```bash
# From project root
docker-compose up frontend
```

## Monorepo

This frontend is part of a monorepo structure:

```
skeleton/
├── backend/               # Go API server
├── frontend/              # This Next.js app
├── shared/               # Shared types
├── docker-compose.yml
└── Makefile
```

Use root Makefile commands:

```bash
# Start all services
make dev

# Start only backend
make dev-backend

# Start only frontend
make dev-frontend
```

## Learn More

- [Next.js Documentation](https://nextjs.org/docs)
- [Tailwind CSS](https://tailwindcss.com/docs)
- [shadcn/ui](https://ui.shadcn.com/)
- [TypeScript](https://www.typescriptlang.org/docs/)