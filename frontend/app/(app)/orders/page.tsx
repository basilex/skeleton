'use client'

import Link from 'next/link'

const orders = [
  { id: '1', number: 'ORD-2026-001', customer: 'Acme Corp', total: 5000, status: 'processing', date: '2026-04-09' },
  { id: '2', number: 'ORD-2026-002', customer: 'Globex Inc', total: 3200, status: 'shipped', date: '2026-04-08' },
  { id: '3', number: 'ORD-2026-003', customer: 'Initech', total: 8900, status: 'delivered', date: '2026-04-07' },
]

export default function OrdersPage() {
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
        <button className="rounded-md bg-secondary px-3 py-1.5 text-xs font-medium">All</button>
        <button className="rounded-md border border-border px-3 py-1.5 text-xs font-medium text-muted-foreground hover:bg-secondary">
          Pending
        </button>
        <button className="rounded-md border border-border px-3 py-1.5 text-xs font-medium text-muted-foreground hover:bg-secondary">
          Processing
        </button>
        <button className="rounded-md border border-border px-3 py-1.5 text-xs font-medium text-muted-foreground hover:bg-secondary">
          Shipped
        </button>
        <button className="rounded-md border border-border px-3 py-1.5 text-xs font-medium text-muted-foreground hover:bg-secondary">
          Delivered
        </button>
      </div>

      {/* Table */}
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
                <td className="px-4 py-3 text-sm text-muted-foreground">{order.customer}</td>
                <td className="px-4 py-3 text-right text-sm">${(order.total / 100).toLocaleString()}</td>
                <td className="px-4 py-3 text-center">
                  <span className={`inline-flex rounded-full px-2 py-1 text-xs font-medium ${
                    order.status === 'processing' ? 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200' :
                    order.status === 'shipped' ? 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200' :
                    'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
                  }`}>
                    {order.status}
                  </span>
                </td>
                <td className="px-4 py-3 text-sm text-muted-foreground">{order.date}</td>
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
    </div>
  )
}