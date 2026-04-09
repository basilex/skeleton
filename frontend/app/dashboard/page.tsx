'use client'

import { useAuth } from '@/lib/auth'
import { RequireAuth } from '@/lib/auth'

export default function DashboardPage() {
  const { user } = useAuth()

  return (
    <RequireAuth>
      <div className="min-h-screen bg-background">
        <header className="border-b border-border">
          <div className="container mx-auto flex h-14 items-center justify-between px-4">
            <h1 className="text-lg font-semibold">Skeleton CRM</h1>
            <div className="flex items-center space-x-4">
              <span className="text-sm text-muted-foreground">
                {user?.email}
              </span>
              <span className="text-xs rounded bg-secondary px-2 py-1">
                {user?.roles?.join(', ')}
              </span>
            </div>
          </div>
        </header>

        <main className="container mx-auto px-4 py-6">
          <div className="space-y-6">
            <div>
              <h2 className="text-xl font-semibold">Dashboard</h2>
              <p className="text-sm text-muted-foreground">
                Welcome back, {user?.name}
              </p>
            </div>

            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
              <div className="rounded-md border border-border p-4">
                <h3 className="text-sm font-medium text-muted-foreground">
                  Total Customers
                </h3>
                <p className="mt-2 text-2xl font-semibold">0</p>
              </div>
              <div className="rounded-md border border-border p-4">
                <h3 className="text-sm font-medium text-muted-foreground">
                  Pending Invoices
                </h3>
                <p className="mt-2 text-2xl font-semibold">0</p>
              </div>
              <div className="rounded-md border border-border p-4">
                <h3 className="text-sm font-medium text-muted-foreground">
                  Active Orders
                </h3>
                <p className="mt-2 text-2xl font-semibold">0</p>
              </div>
              <div className="rounded-md border border-border p-4">
                <h3 className="text-sm font-medium text-muted-foreground">
                  Low Stock Items
                </h3>
                <p className="mt-2 text-2xl font-semibold">0</p>
              </div>
            </div>

            <div className="rounded-md border border-border p-4">
              <h3 className="text-sm font-medium mb-3">Quick Actions</h3>
              <div className="grid gap-2 md:grid-cols-4">
                <button className="rounded border border-border bg-background px-4 py-2 text-sm hover:bg-secondary text-left">
                  + New Customer
                </button>
                <button className="rounded border border-border bg-background px-4 py-2 text-sm hover:bg-secondary text-left">
                  + New Invoice
                </button>
                <button className="rounded border border-border bg-background px-4 py-2 text-sm hover:bg-secondary text-left">
                  + New Order
                </button>
                <button className="rounded border border-border bg-background px-4 py-2 text-sm hover:bg-secondary text-left">
                  + New Product
                </button>
              </div>
            </div>
          </div>
        </main>
      </div>
    </RequireAuth>
  )
}