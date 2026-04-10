'use client'

import Link from 'next/link'
import { useState } from 'react'
import { useOrders } from '@/lib/query'

export default function OrdersPage() {
  const [statusFilter, setStatusFilter] = useState<string | null>(null)
  const { data, isLoading, error } = useOrders({ 
    status: statusFilter || undefined, 
    limit: 100 
  })

  const orders = data?.items ?? []

  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Orders</h1>
          <p className="text-sm text-muted-foreground">
            Track and manage orders
          </p>
        </div>
        <Link
          href="/orders/new"
          className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
        >
          + Create Order
        </Link>
      </div>

      {/* Filters */}
      <div className="mb-4 flex gap-2">
        <button 
          onClick={() => setStatusFilter(null)}
          className={`rounded-md px-3 py-1.5 text-xs font-medium ${
            statusFilter === null 
              ? 'bg-secondary' 
              : 'border border-border text-muted-foreground hover:bg-secondary'
          }`}>
          All
        </button>
        <button 
          onClick={() => setStatusFilter('pending')}
          className={`rounded-md px-3 py-1.5 text-xs font-medium ${
            statusFilter === 'pending' 
              ? 'bg-secondary' 
              : 'border border-border text-muted-foreground hover:bg-secondary'
          }`}>
          Pending
        </button>
        <button 
          onClick={() => setStatusFilter('processing')}
          className={`rounded-md px-3 py-1.5 text-xs font-medium ${
            statusFilter === 'processing' 
              ? 'bg-secondary' 
              : 'border border-border text-muted-foreground hover:bg-secondary'
          }`}>
          Processing
        </button>
        <button 
          onClick={() => setStatusFilter('shipped')}
          className={`rounded-md px-3 py-1.5 text-xs font-medium ${
            statusFilter === 'shipped' 
              ? 'bg-secondary' 
              : 'border border-border text-muted-foreground hover:bg-secondary'
          }`}>
          Shipped
        </button>
        <button 
          onClick={() => setStatusFilter('delivered')}
          className={`rounded-md px-3 py-1.5 text-xs font-medium ${
            statusFilter === 'delivered' 
              ? 'bg-secondary' 
              : 'border border-border text-muted-foreground hover:bg-secondary'
          }`}>
          Delivered
        </button>
      </div>

      {/* Loading State */}
      {isLoading && (
        <div className="rounded-md border border-border p-8 text-center text-sm text-muted-foreground">
          Loading orders...
        </div>
      )}

      {/* Error State */}
      {error && (
        <div className="rounded-md border border-border p-8 text-center text-sm text-red-600">
          Failed to load orders. Please try again.
        </div>
      )}

      {/* Empty State */}
      {!isLoading && !error && orders.length === 0 && (
        <div className="rounded-md border border-border p-8 text-center text-sm text-muted-foreground">
          No orders found.
        </div>
      )}

      {/* Table */}
      {!isLoading && !error && orders.length > 0 && (
        <div className="rounded-md border border-border">
          <table className="w-full">
            <thead className="border-b border-border bg-muted/50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">Number</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">Customer</th>
                <th className="px-4 py-3 text-right text-xs font-medium text-muted-foreground">Total</th>
                <th className="px-4 py-3 text-center text-xs font-medium text-muted-foreground">Status</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">Date</th>
                <th className="px-4 py-3 text-right text-xs font-medium text-muted-foreground">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {orders.map((order) => (
                <tr key={order.id} className="hover:bg-muted/50">
                  <td className="px-4 py-3">
                    <Link href={`/orders/${order.id}`} className="font-medium hover:underline">
                      {order.number}
                    </Link>
                  </td>
                  <td className="px-4 py-3 text-sm text-muted-foreground">{order.customer_name}</td>
                  <td className="px-4 py-3 text-right text-sm">${order.total.toLocaleString()}</td>
                  <td className="px-4 py-3 text-center">
                    <span className={`inline-flex rounded-full px-2 py-1 text-xs font-medium capitalize ${
                      order.status === 'processing' ? 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200' :
                      order.status === 'shipped' ? 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200' :
                      order.status === 'delivered' ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200' :
                      'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200'
                    }`}>
                      {order.status}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-sm text-muted-foreground">
                    {new Date(order.created_at).toLocaleDateString()}
                  </td>
                  <td className="px-4 py-3 text-right">
                    <Link href={`/orders/${order.id}`} className="text-sm text-muted-foreground hover:text-foreground">
                      View
                    </Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}