'use client'

import { useAuth } from '@/lib/auth'
import Link from 'next/link'

export default function DashboardPage() {
  const { user } = useAuth()

  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-semibold">Dashboard</h1>
        <p className="text-sm text-muted-foreground">
          Welcome back, {user?.name}
        </p>
      </div>

      {/* Stats */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Link
          href="/customers"
          className="rounded-md border border-border bg-background p-4 hover:bg-secondary/50"
        >
          <h3 className="text-sm font-medium text-muted-foreground">Total Customers</h3>
          <p className="mt-2 text-2xl font-semibold">248</p>
          <p className="mt-1 text-xs text-muted-foreground">+12 this month</p>
        </Link>

        <Link
          href="/invoices"
          className="rounded-md border border-border bg-background p-4 hover:bg-secondary/50"
        >
          <h3 className="text-sm font-medium text-muted-foreground">Pending Invoices</h3>
          <p className="mt-2 text-2xl font-semibold">23</p>
          <p className="mt-1 text-xs text-muted-foreground">$45,600 outstanding</p>
        </Link>

        <Link
          href="/orders"
          className="rounded-md border border-border bg-background p-4 hover:bg-secondary/50"
        >
          <h3 className="text-sm font-medium text-muted-foreground">Active Orders</h3>
          <p className="mt-2 text-2xl font-semibold">17</p>
          <p className="mt-1 text-xs text-muted-foreground">5 shipping today</p>
        </Link>

        <Link
          href="/inventory"
          className="rounded-md border border-border bg-background p-4 hover:bg-secondary/50"
        >
          <h3 className="text-sm font-medium text-muted-foreground">Low Stock Items</h3>
          <p className="mt-2 text-2xl font-semibold">8</p>
          <p className="mt-1 text-xs text-muted-foreground">Needs attention</p>
        </Link>
      </div>

      {/* Quick Actions */}
      <div className="mt-6 rounded-md border border-border bg-background p-4">
        <h3 className="mb-3 text-sm font-medium">Quick Actions</h3>
        <div className="grid gap-2 md:grid-cols-4">
          <Link
            href="/customers/new"
            className="rounded border border-border bg-background px-4 py-2 text-sm hover:bg-secondary"
          >
            + New Customer
          </Link>
          <Link
            href="/invoices/new"
            className="rounded border border-border bg-background px-4 py-2 text-sm hover:bg-secondary"
          >
            + New Invoice
          </Link>
          <Link
            href="/orders/new"
            className="rounded border border-border bg-background px-4 py-2 text-sm hover:bg-secondary"
          >
            + New Order
          </Link>
          <Link
            href="/products/new"
            className="rounded border border-border bg-background px-4 py-2 text-sm hover:bg-secondary"
          >
            + New Product
          </Link>
        </div>
      </div>

      {/* Recent Activity */}
      <div className="mt-6 grid gap-6 md:grid-cols-2">
        {/* Recent Invoices */}
        <div className="rounded-md border border-border bg-background">
          <div className="border-b border-border p-4">
            <h3 className="font-medium">Recent Invoices</h3>
          </div>
          <div className="divide-y divide-border">
            {[1, 2, 3, 4, 5].map((i) => (
              <div key={i} className="flex items-center justify-between p-4">
                <div>
                  <p className="text-sm font-medium">INV-2026-{i.toString().padStart(3, '0')}</p>
                  <p className="text-xs text-muted-foreground">Customer {i}</p>
                </div>
                <div className="text-right">
                  <p className="text-sm font-medium">$1,250</p>
                  <p className="text-xs text-muted-foreground">Pending</p>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Recent Orders */}
        <div className="rounded-md border border-border bg-background">
          <div className="border-b border-border p-4">
            <h3 className="font-medium">Recent Orders</h3>
          </div>
          <div className="divide-y divide-border">
            {[1, 2, 3, 4, 5].map((i) => (
              <div key={i} className="flex items-center justify-between p-4">
                <div>
                  <p className="text-sm font-medium">ORD-2026-{i.toString().padStart(3, '0')}</p>
                  <p className="text-xs text-muted-foreground">Customer {i}</p>
                </div>
                <div className="text-right">
                  <p className="text-sm font-medium">$890</p>
                  <p className="text-xs text-muted-foreground">Processing</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}