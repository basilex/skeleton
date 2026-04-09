import { apiClient } from './client';
import type {
  User,
  Customer,
  Supplier,
  Account,
  Transaction,
  Invoice,
  Payment,
  Order,
  Stock,
  Product,
  PaginatedResponse,
  PaginationParams,
} from '@shared/types/api';

// ============================================
// Auth Endpoints
// ============================================

export const authAPI = {
  login: (email: string, password: string) =>
    apiClient.post<{ token: string; user: User }>('/auth/login', { email, password }),

  register: (email: string, password: string) =>
    apiClient.post<{ token: string; user: User }>('/auth/register', { email, password }),

  logout: () =>
    apiClient.post('/auth/logout'),

  me: () =>
    apiClient.get<User>('/auth/me'),
};

// ============================================
// Parties Endpoints
// ============================================

export const partiesAPI = {
  getCustomers: (params?: PaginationParams) =>
    apiClient.get<PaginatedResponse<Customer>>(`/customers?page=${params?.page || 1}&limit=${params?.limit || 20}`),

  getCustomer: (id: string) =>
    apiClient.get<Customer>(`/customers/${id}`),

  createCustomer: (data: Partial<Customer>) =>
    apiClient.post<Customer>('/customers', data),

  updateCustomer: (id: string, data: Partial<Customer>) =>
    apiClient.put<Customer>(`/customers/${id}`, data),

  deleteCustomer: (id: string) =>
    apiClient.delete(`/customers/${id}`),

  getSuppliers: (params?: PaginationParams) =>
    apiClient.get<PaginatedResponse<Supplier>>(`/suppliers?page=${params?.page || 1}&limit=${params?.limit || 20}`),

  getSupplier: (id: string) =>
    apiClient.get<Supplier>(`/suppliers/${id}`),

  createSupplier: (data: Partial<Supplier>) =>
    apiClient.post<Supplier>('/suppliers', data),
};

// ============================================
// Accounting Endpoints
// ============================================

export const accountingAPI = {
  getAccounts: (params?: PaginationParams) =>
    apiClient.get<PaginatedResponse<Account>>(`/accounts?page=${params?.page || 1}&limit=${params?.limit || 20}`),

  getAccount: (id: string) =>
    apiClient.get<Account>(`/accounts/${id}`),

  createAccount: (data: Partial<Account>) =>
    apiClient.post<Account>('/accounts', data),

  getTransactions: (accountId?: string) =>
    apiClient.get<Transaction[]>(accountId ? `/accounts/${accountId}/transactions` : '/transactions'),

  createTransaction: (data: Partial<Transaction>) =>
    apiClient.post<Transaction>('/transactions', data),
};

// ============================================
// Invoicing Endpoints
// ============================================

export const invoicingAPI = {
  getInvoices: (params?: PaginationParams) =>
    apiClient.get<PaginatedResponse<Invoice>>(`/invoices?page=${params?.page || 1}&limit=${params?.limit || 20}`),

  getInvoice: (id: string) =>
    apiClient.get<Invoice>(`/invoices/${id}`),

  createInvoice: (data: Partial<Invoice>) =>
    apiClient.post<Invoice>('/invoices', data),

  updateInvoice: (id: string, data: Partial<Invoice>) =>
    apiClient.put<Invoice>(`/invoices/${id}`, data),

  sendInvoice: (id: string) =>
    apiClient.post<Invoice>(`/invoices/${id}/send`),

  getPayments: (invoiceId: string) =>
    apiClient.get<Payment[]>(`/invoices/${invoiceId}/payments`),

  createPayment: (invoiceId: string, data: Partial<Payment>) =>
    apiClient.post<Payment>(`/invoices/${invoiceId}/payments`, data),
};

// ============================================
// Ordering Endpoints
// ============================================

export const orderingAPI = {
  getOrders: (params?: PaginationParams) =>
    apiClient.get<PaginatedResponse<Order>>(`/orders?page=${params?.page || 1}&limit=${params?.limit || 20}`),

  getOrder: (id: string) =>
    apiClient.get<Order>(`/orders/${id}`),

  createOrder: (data: Partial<Order>) =>
    apiClient.post<Order>('/orders', data),

  updateOrder: (id: string, data: Partial<Order>) =>
    apiClient.put<Order>(`/orders/${id}`, data),

  confirmOrder: (id: string) =>
    apiClient.post<Order>(`/orders/${id}/confirm`),

  cancelOrder: (id: string) =>
    apiClient.post<Order>(`/orders/${id}/cancel`),
};

// ============================================
// Inventory Endpoints
// ============================================

export const inventoryAPI = {
  getStock: (warehouseId?: string) =>
    apiClient.get<Stock[]>(warehouseId ? `/warehouses/${warehouseId}/stock` : '/stock'),

  updateStock: (stockId: string, quantity: number) =>
    apiClient.patch<Stock>(`/stock/${stockId}`, { quantity }),
};

// ============================================
// Catalog Endpoints
// ============================================

export const catalogAPI = {
  getProducts: (params?: PaginationParams) =>
    apiClient.get<PaginatedResponse<Product>>(`/products?page=${params?.page || 1}&limit=${params?.limit || 20}`),

  getProduct: (id: string) =>
    apiClient.get<Product>(`/products/${id}`),

  createProduct: (data: Partial<Product>) =>
    apiClient.post<Product>('/products', data),

  updateProduct: (id: string, data: Partial<Product>) =>
    apiClient.put<Product>(`/products/${id}`, data),

  deleteProduct: (id: string) =>
    apiClient.delete(`/products/${id}`),
};