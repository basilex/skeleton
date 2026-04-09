/**
 * API Types for frontend-to-backend communication
 * 
 * These types match the Go domain models in backend/internal/{context}/domain
 */

import { Money } from './money';

// ============================================
// Identity Context
// ============================================

export interface User {
  id: string;
  email: string;
  role: string;
  status: 'active' | 'inactive' | 'suspended';
  createdAt: string;
  updatedAt: string;
}

export interface Session {
  id: string;
  userId: string;
  token: string;
  expiresAt: string;
  createdAt: string;
}

// ============================================
// Parties Context  
// ============================================

export interface Customer {
  id: string;
  name: string;
  taxId: string;
  totalPurchases: Money;
  creditLimit: Money;
  currentCredit: Money;
  status: 'active' | 'inactive' | 'blacklisted';
  loyaltyLevel: 'bronze' | 'silver' | 'gold' | 'platinum';
  createdAt: string;
}

export interface Supplier {
  id: string;
  name: string;
  taxId: string;
  status: 'active' | 'inactive';
  contracts: string[];
  rating: number;
}

// ============================================
// Accounting Context
// ============================================

export type AccountType = 'asset' | 'liability' | 'equity' | 'revenue' | 'expense';

export interface Account {
  id: string;
  code: string;
  name: string;
  type: AccountType;
  currency: string;
  balance: Money;
  status: 'active' | 'inactive';
  parentId?: string;
  createdAt: string;
}

export interface Transaction {
  id: string;
  fromAccountId: string;
  toAccountId: string;
  amount: Money;
  description: string;
  status: 'pending' | 'completed' | 'failed';
  createdAt: string;
}

// ============================================
// Invoicing Context
// ============================================

export type InvoiceStatus = 'draft' | 'sent' | 'viewed' | 'paid' | 'overdue' | 'cancelled';

export interface Invoice {
  id: string;
  invoiceNumber: string;
  customerId: string;
  subtotal: Money;
  taxRate: number;
  taxAmount: Money;
  discount: Money;
  total: Money;
  paidAmount: Money;
  status: InvoiceStatus;
  dueDate: string;
  createdAt: string;
}

export interface InvoiceLine {
  id: string;
  description: string;
  quantity: number;
  unit: string;
  unitPrice: Money;
  discount: Money;
  total: Money;
}

export interface Payment {
  id: string;
  invoiceId: string;
  amount: Money;
  method: 'card' | 'bank_transfer' | 'cash' | 'check';
  status: 'pending' | 'completed' | 'failed';
  createdAt: string;
}

// ============================================
// Ordering Context
// ============================================

export type OrderStatus = 'draft' | 'confirmed' | 'processing' | 'completed' | 'cancelled';

export interface Order {
  id: string;
  orderNumber: string;
  customerId: string;
  supplierId: string;
  subtotal: Money;
  taxAmount: Money;
  discount: Money;
  total: Money;
  status: OrderStatus;
  orderDate: string;
  createdAt: string;
}

export interface OrderLine {
  id: string;
  itemId: string;
  itemName: string;
  quantity: number;
  unit: string;
  unitPrice: Money;
  discount: Money;
  total: Money;
}

// ============================================
// Inventory Context
// ============================================

export interface Stock {
  id: string;
  warehouseId: string;
  itemId: string;
  quantity: number;
  reservedQuantity: number;
  availableQuantity: number;
  unit: string;
  status: 'active' | 'inactive';
}

export interface StockAdjustment {
  id: string;
  stockId: string;
  quantity: number;
  reason: string;
  type: 'gain' | 'loss';
  createdAt: string;
}

// ============================================
// Catalog Context
// ============================================

export interface Product {
  id: string;
  sku: string;
  name: string;
  description: string;
  categoryId: string;
  basePrice: Money;
  status: 'active' | 'inactive' | 'discontinued';
  variants: ProductVariant[];
}

export interface ProductVariant {
  id: string;
  productId: string;
  sku: string;
  attributes: Record<string, string>;
  price: Money;
  stock: number;
}

// ============================================
// Pagination
// ============================================

export interface PaginationParams {
  page: number;
  limit: number;
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  limit: number;
  hasMore: boolean;
}