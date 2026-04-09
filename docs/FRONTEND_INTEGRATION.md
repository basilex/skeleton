# Frontend Integration Guide: Next.js + shadcn/ui

## Overview

This document describes how to integrate Next.js frontend with shadcn/ui components with the Skeleton Business Engine API.

## API Architecture (Already Frontend-Ready ✅)

### 1. CORS Configuration ✅

```go
// cmd/api/routes.go
func corsMiddleware() gin.HandlerFunc {
    return cors.New(cors.Config{
        AllowOrigins:     []string{"http://localhost:3000", "https://yourdomain.com"},
        AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
        ExposeHeaders:    []string{"Content-Length", "X-Request-Id"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    })
}
```

**For Next.js Development:**
- AllowOrigins should include `http://localhost:3000`
- AllowCredentials enables cookies/sessions

### 2. Standardized Error Responses ✅

All endpoints return **RFC 7807 Problem Details** format:

```typescript
// TypeScript interface for API errors
interface APIError {
  type: string;      // "https://skeleton.app/errors/validation"
  title: string;     // "Validation Error"
  status: number;   // 422
  detail: string;   // "Invalid input data"
  instance: string; // "/api/v1/warehouses"
  request_id: string; // "req_123abc"
}
```

**Example Error Response:**
```json
{
  "type": "https://skeleton.app/errors/validation",
  "title": "Validation Error",
  "status": 422,
  "detail": "warehouse name cannot be empty",
  "instance": "/api/v1/warehouses",
  "request_id": "req_01HXYZ123"
}
```

### 3. Pagination Format ✅

**Cursor-based pagination for all list endpoints:**

```typescript
// Request (query parameters)
interface PaginationQuery {
  cursor?: string;  // Next page cursor
  limit?: number;   // Items per page (default: 20)
}

// Response
interface PaginatedResponse<T> {
  items: T[];
  next_cursor: string | null;
}

// Example: GET /api/v1/warehouses?cursor=abc123&limit=20
// Response:
{
  "warehouses": [
    { "id": "...", "name": "Warehouse 1", ... },
    { "id": "...", "name": "Warehouse 2", ... }
  ],
  "next_cursor": "def456" // null if last page
}
```

### 4. Request/Response Formats ✅

**All endpoints use JSON:**
- Content-Type: `application/json`
- Dates: ISO 8601 format (`"2024-01-08T15:04:05Z07:00"`)
- UUIDs: String format (`"0194a7b2-1234-5678-9abc-def012345678"`)
- Money: Float with 2 decimals (`150.99`)
- Enums: String values (`"active"`, `"inactive"`)

## Inventory Context API Endpoints

### Authentication

All inventory endpoints require:
```typescript
// Headers
{
  "Authorization": "Bearer <jwt_token>",
  "Content-Type": "application/json"
}
```

### Warehouses API

```typescript
// GET /api/v1/warehouses - List warehouses
interface ListWarehousesQuery {
  status?: 'active' | 'inactive' | 'maintenance';
  code?: string;
  cursor?: string;
  limit?: number;
}

interface Warehouse {
  id: string;
  name: string;
  code: string;
  location: string;
  capacity: number;
  status: 'active' | 'inactive' | 'maintenance';
  metadata: Record<string, string>;
  created_at: string;
  updated_at: string;
}

// POST /api/v1/warehouses - Create warehouse
// Request body:
{
  "name": "Main Warehouse",
  "code": "WH-001",
  "location": "New York"
}

// PUT /api/v1/warehouses/:id - Update warehouse
// Request body:
{
  "name": "Updated Name",
  "capacity": 10000,
  "activate": true
}
```

### Stock API

```typescript
// GET /api/v1/stock - List stock
interface ListStockQuery {
  item_id?: string;
  warehouse_id?: string;
  available?: boolean;
  cursor?: string;
  limit?: number;
}

interface Stock {
  id: string;
  item_id: string;
  warehouse_id: string;
  quantity: number;
  reserved_qty: number;
  available_qty: number;
  reorder_point: number;
  reorder_quantity: number;
  last_movement_id: string;
  created_at: string;
  updated_at: string;
}

// POST /api/v1/stock - Create stock
// POST /api/v1/stock/:id/adjust - Adjust stock
// POST /api/v1/stock/receipt - Receipt stock
// POST /api/v1/stock/issue - Issue stock
// POST /api/v1/stock/transfer - Transfer stock
// POST /api/v1/stock/reserve - Reserve stock
```

### Stock Movements API

```typescript
// GET /api/v1/movements - List movements
interface ListMovementsQuery {
  item_id?: string;
  warehouse_id?: string;
  movement_type?: 'receipt' | 'issue' | 'transfer' | 'adjustment' | 'return';
  reference_type?: string;
  start_date?: string;
  end_date?: string;
  cursor?: string;
  limit?: number;
}

interface StockMovement {
  id: string;
  movement_type: 'receipt' | 'issue' | 'transfer' | 'adjustment' | 'return';
  item_id: string;
  from_warehouse: string;
  to_warehouse: string;
  quantity: number;
  reference_id: string;
  reference_type: string;
  notes: string;
  occurred_at: string;
  created_at: string;
}
```

### Reservations API

```typescript
// GET /api/v1/reservations - List reservations
interface ListReservationsQuery {
  order_id: string; // Required
}

interface StockReservation {
  id: string;
  stock_id: string;
  order_id: string;
  quantity: number;
  status: 'active' | 'fulfilled' | 'cancelled' | 'expired';
  reserved_at: string;
  expires_at: string | null;
  fulfilled_at: string | null;
  cancelled_at: string | null;
  created_at: string;
  updated_at: string;
}

// POST /api/v1/reservations/fulfill - Fulfill reservation
// POST /api/v1/reservations/cancel - Cancel reservation
```

## Next.js Setup

### 1. API Client with Type Safety

```bash
# Generate TypeScript types from Swagger
npm install openapi-typescript-codegen
```

```typescript
// lib/api-client.ts
import axios from 'axios';

const apiClient = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true, // For cookies
});

// Add auth token to requests
apiClient.interceptors.request.use((config) => {
  const token = getAuthToken(); // From cookies or localStorage
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Handle API errors
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.data) {
      const apiError: APIError = error.response.data;
      throw new APIError(apiError.type, apiError.title, apiError.status, apiError.detail);
    }
    throw error;
  }
);

export default apiClient;
```

### 2. React Query Hooks

```typescript
// hooks/use-warehouses.ts
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import apiClient from '@/lib/api-client';

// List warehouses
export function useWarehouses(query?: ListWarehousesQuery) {
  return useQuery({
    queryKey: ['warehouses', query],
    queryFn: async () => {
      const { data } = await apiClient.get('/warehouses', { params: query });
      return data;
    },
  });
}

// Get single warehouse
export function useWarehouse(id: string) {
  return useQuery({
    queryKey: ['warehouse', id],
    queryFn: async () => {
      const { data } = await apiClient.get(`/warehouses/${id}`);
      return data;
    },
    enabled: !!id,
  });
}

// Create warehouse
export function useCreateWarehouse() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (warehouse: CreateWarehouseRequest) => {
      const { data } = await apiClient.post('/warehouses', warehouse);
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['warehouses'] });
    },
  });
}
```

### 3. shadcn/ui Components

```typescript
// components/warehouses/warehouse-form.tsx
import { useForm } from 'react-hook-form';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useCreateWarehouse } from '@/hooks/use-warehouses';

export function WarehouseForm() {
  const { register, handleSubmit } = useForm<CreateWarehouseRequest>();
  const createWarehouse = useCreateWarehouse();

  const onSubmit = (data: CreateWarehouseRequest) => {
    createWarehouse.mutate(data);
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      <div>
        <Label htmlFor="name">Warehouse Name</Label>
        <Input
          id="name"
          {...register('name', { required: true })}
          placeholder="Main Warehouse"
        />
      </div>
      
      <div>
        <Label htmlFor="code">Code</Label>
        <Input
          id="code"
          {...register('code')}
          placeholder="WH-001"
        />
      </div>
      
      <Button type="submit" loading={createWarehouse.isPending}>
        Create Warehouse
      </Button>
    </form>
  );
}
```

### 4. Cursor Pagination Component

```typescript
// components/ui/cursor-pagination.tsx
import { Button } from '@/components/ui/button';

interface CursorPaginationProps {
  nextCursor: string | null;
  onLoadMore: (cursor: string) => void;
  isLoading: boolean;
}

export function CursorPagination({ nextCursor, onLoadMore, isLoading }: CursorPaginationProps) {
  if (!nextCursor) return null;

  return (
    <div className="flex justify-center py-4">
      <Button
        variant="outline"
        onClick={() => onLoadMore(nextCursor)}
        disabled={isLoading}
      >
        {isLoading ? 'Loading...' : 'Load More'}
      </Button>
    </div>
  );
}
```

### 5. Error Handling Component

```typescript
// components/ui/api-error.tsx
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { AlertCircle } from 'lucide-react';

interface APIErrorProps {
  error: APIError | null;
}

export function APIErrorAlert({ error }: APIErrorProps) {
  if (!error) return null;

  return (
    <Alert variant="destructive">
      <AlertCircle className="h-4 w-4" />
      <AlertTitle>{error.title}</AlertTitle>
      <AlertDescription>{error.detail}</AlertDescription>
    </Alert>
  );
}
```

## Data Table Example

```typescript
// components/warehouses/warehouses-table.tsx
import { useWarehouses } from '@/hooks/use-warehouses';
import { DataTable } from '@/components/ui/data-table';
import { columns } from './columns';

export function WarehousesTable() {
  const [cursor, setCursor] = useState<string | undefined>();
  
  const { data, isLoading } = useWarehouses({ 
    cursor,
    limit: 20,
  });

  return (
    <div>
      <DataTable
        columns={columns}
        data={data?.warehouses || []}
        isLoading={isLoading}
      />
      
      {data?.next_cursor && (
        <CursorPagination
          nextCursor={data.next_cursor}
          onLoadMore={setCursor}
          isLoading={isLoading}
        />
      )}
    </div>
  );
}
```

## Required Changes to Backend

### Already Implemented ✅

1. **CORS** - Already configured in `cmd/api/routes.go`
2. **Swagger Annotations** - All endpoints have `@Summary`, `@Description`, etc.
3. **Standardized Errors** - RFC 7807 format via `pkg/apierror`
4. **Cursor Pagination** - All list endpoints support cursor-based pagination
5. **JSON Responses** - All endpoints return JSON
6. **Request ID** - Every request has unique `X-Request-Id` header

### Recommended Additions

```go
// Add to cmd/api/main.go or routes.go

// Health check for Next.js
v1.GET("/health", func(c *gin.Context) {
    c.JSON(200, gin.H{
        "status": "ok",
        "timestamp": time.Now().Format(time.RFC3339),
    })
})

// API info endpoint
v1.GET("/", func(c *gin.Context) {
    c.JSON(200, gin.H{
        "name": "Skeleton Business Engine",
        "version": "1.0.0",
        "contexts": []string{
            "identity", "parties", "contracts", "accounting",
            "ordering", "catalog", "invoicing", "documents", "inventory",
        },
    })
})
```

## TypeScript Types Generation

```bash
# Generate types from Swagger
npx openapi-typescript-codegen \
  --input http://localhost:8080/swagger/doc.json \
  --output ./src/types/api \
  --client axios
```

This will generate:
- `src/types/api/models/Warehouse.ts`
- `src/types/api/models/Stock.ts`
- `src/types/api/models/StockMovement.ts`
- `src/types/api/models/StockReservation.ts`
- `src/types/api/services/WarehousesService.ts`
- `src/types/api/services/StockService.ts`

## Permissions & RBAC

All inventory endpoints use RBAC middleware:
- `inventory:read` - Required for GET endpoints
- `inventory:write` - Required for POST/PUT/PATCH/DELETE endpoints

```typescript
// Check permissions in Next.js
const userPermissions = ['inventory:read', 'inventory:write'];

// Component guard
if (!userPermissions.includes('inventory:read')) {
  return <AccessDenied />;
}
```

## Environment Variables

```env
# .env.local (Next.js)
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
NEXT_PUBLIC_APP_NAME=Skeleton Business Engine

# Production
NEXT_PUBLIC_API_URL=https://api.yourdomain.com/api/v1
```

## Summary

✅ **Backend Ready for Next.js:**
- CORS configured
- Standardized error responses (RFC 7807)
- Cursor-based pagination
- Swagger annotations for type generation
- JWT authentication
- RBAC permissions
- Request ID tracking

✅ **Frontend Requirements:**
- Use React Query for data fetching
- Implement cursor pagination UI
- Generate TypeScript types from Swagger
- Handle API errors consistently
- Use shadcn/ui components for forms/tables

**No backend changes needed!** The Inventory Context is fully ready for Next.js integration.