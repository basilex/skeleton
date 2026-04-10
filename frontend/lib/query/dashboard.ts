'use client'

import { useQuery } from '@tanstack/react-query'
import { useCustomers } from './customers'
import { useInvoices } from './invoices'
import { useOrders } from './orders'

export interface DashboardStats {
  totalCustomers: number
  pendingInvoices: number
  activeOrders: number
  lowStockItems: number
  outstandingAmount: number
}

export function useDashboardStats() {
  const customersQuery = useCustomers({ limit: 1 })
  const invoicesQuery = useInvoices({ status: 'pending', limit: 100 })
  const ordersQuery = useOrders({ status: 'processing', limit: 100 })

  const isLoading = 
    customersQuery.isLoading || 
    invoicesQuery.isLoading || 
    ordersQuery.isLoading

  const error = 
    customersQuery.error || 
    invoicesQuery.error || 
    ordersQuery.error

  const stats: DashboardStats = {
    totalCustomers: customersQuery.data?.items.length ?? 0,
    pendingInvoices: invoicesQuery.data?.items.length ?? 0,
    activeOrders: ordersQuery.data?.items.length ?? 0,
    lowStockItems: 0, // Will need inventory API
    outstandingAmount: 0, // Will need to calculate from invoices
  }

  return {
    stats,
    isLoading,
    error,
    recentInvoices: invoicesQuery.data?.items.slice(0, 5) ?? [],
    recentOrders: ordersQuery.data?.items.slice(0, 5) ?? [],
  }
}