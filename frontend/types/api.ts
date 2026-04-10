export interface Address {
  street: string
  city: string
  region: string
  postal_code: string
  country: string
}

export interface BankAccount {
  bank_name: string
  account_name: string
  account_number: string
  swift_code: string
  iban: string
  currency: string
}

export interface Customer {
  id: string
  name: string
  tax_id: string
  email: string
  phone: string
  address: Address
  website: string
  social_media: Record<string, string>
  bank_account: BankAccount
  status: 'active' | 'inactive' | 'blacklisted'
  loyalty_level: string
  total_purchases: number
  created_at: string
  updated_at: string
}

export interface Supplier {
  id: string
  name: string
  tax_id: string
  email: string
  phone: string
  address: Address
  website: string
  social_media: Record<string, string>
  bank_account: BankAccount
  status: 'active' | 'inactive' | 'blacklisted'
  rating: number
  total_orders: number
  created_at: string
  updated_at: string
}

export interface InvoiceLine {
  id: string
  description: string
  quantity: number
  unit_price: number
  total: number
}

export interface Invoice {
  id: string
  number: string
  customer_id: string
  customer_name: string
  status: 'draft' | 'pending' | 'paid' | 'overdue' | 'cancelled'
  issue_date: string
  due_date: string
  subtotal: number
  tax: number
  total: number
  notes: string
  lines: InvoiceLine[]
  created_at: string
  updated_at: string
}

export interface OrderLine {
  id: string
  product_id: string
  product_name: string
  quantity: number
  unit_price: number
  total: number
}

export interface Order {
  id: string
  number: string
  customer_id: string
  customer_name: string
  status: 'pending' | 'processing' | 'shipped' | 'delivered' | 'cancelled'
  total: number
  lines: OrderLine[]
  created_at: string
  updated_at: string
}

export interface Product {
  id: string
  sku: string
  name: string
  description: string
  category: string
  price: number
  stock: number
  status: 'active' | 'inactive'
}

export interface PageResult<T> {
  items: T[]
  next_cursor?: string
  has_more: boolean
  limit: number
}