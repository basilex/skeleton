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
// Documents Context
// ============================================

export type DocumentStatus = 'draft' | 'pending_approval' | 'approved' | 'rejected' | 'archived';
export type DocumentType = 'invoice' | 'order' | 'quote' | 'report' | 'contract' | 'other';

export interface Document {
  id: string;
  documentNumber: string;
  documentType: DocumentType;
  referenceId: string;
  fileId?: string;
  status: DocumentStatus;
  metadata: Record<string, string>;
  signatures: Signature[];
  versions: DocumentVersion[];
  currentVersion: number;
  createdAt: string;
  updatedAt: string;
}

export interface Signature {
  id: string;
  documentId: string;
  signatoryId: string;
  signatoryName: string;
  signedAt?: string;
  status: 'pending' | 'signed' | 'rejected';
}

export interface DocumentVersion {
  id: string;
  documentId: string;
  versionNumber: number;
  fileId: string;
  createdBy: string;
  createdAt: string;
}

// ============================================
// Tasks Context
// ============================================

export type TaskType = 
  | 'send_email'
  | 'send_sms'
  | 'send_push'
  | 'send_in_app'
  | 'process_file'
  | 'generate_thumbnail'
  | 'cleanup_old_data'
  | 'generate_report';

export type TaskStatus = 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';

export interface Task {
  id: string;
  taskType: TaskType;
  payload: Record<string, unknown>;
  status: TaskStatus;
  priority: number;
  retries: number;
  maxRetries: number;
  scheduledAt: string;
  executedAt?: string;
  completedAt?: string;
  errorMessage?: string;
  createdAt: string;
  updatedAt: string;
}

// ============================================
// Notifications Context
// ============================================

export type NotificationType = 'email' | 'sms' | 'push' | 'in_app';
export type NotificationStatus = 'pending' | 'sent' | 'delivered' | 'failed';

export interface Notification {
  id: string;
  userId: string;
  type: NotificationType;
  title: string;
  message: string;
  data?: Record<string, unknown>;
  status: NotificationStatus;
  read: boolean;
  sentAt?: string;
  readAt?: string;
  createdAt: string;
}

// ============================================
// Audit Context
// ============================================

export type AuditAction = 'create' | 'read' | 'update' | 'delete' | 'login' | 'logout' | 'assign_role' | 'revoke_role' | 'register';

export interface AuditRecord {
  id: string;
  actorId: string;
  actorType: 'user' | 'system';
  action: AuditAction;
  resourceType: string;
  resourceId: string;
  oldValues?: Record<string, unknown>;
  newValues?: Record<string, unknown>;
  metadata: Record<string, string>;
  ipAddress?: string;
  userAgent?: string;
  createdAt: string;
}

// ============================================
// Files Context
// ============================================

export type FileStatus = 'uploading' | 'processing' | 'ready' | 'error' | 'expired';

export interface File {
  id: string;
  name: string;
  path: string;
  mimeType: string;
  size: number;
  hash: string;
  status: FileStatus;
  ownerId?: string;
  expiresAt?: string;
  metadata: Record<string, string>;
  createdAt: string;
  updatedAt: string;
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