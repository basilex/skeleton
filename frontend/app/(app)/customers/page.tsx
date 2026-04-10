'use client'

import Link from 'next/link'
import { useState } from 'react'
import { useCustomers } from '@/lib/query'

export default function CustomersPage() {
  const [search, setSearch] = useState('')
  const { data, isLoading, error } = useCustomers({ search: search || undefined })

  const customers = data?.items ?? []

  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Customers</h1>
          <p className="text-sm text-muted-foreground">
            Manage your customer relationships
          </p>
        </div>
        <Link
          href="/customers/new"
          className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
        >
          + Add Customer
        </Link>
      </div>

      {/* Search */}
      <div className="mb-4">
        <input
          type="text"
          placeholder="Search customers..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="h-9 w-full rounded-md border border-border bg-background px-3 text-sm"
        />
      </div>

      {/* Loading State */}
      {isLoading && (
        <div className="rounded-md border border-border p-8 text-center text-sm text-muted-foreground">
          Loading customers...
        </div>
      )}

      {/* Error State */}
      {error && (
        <div className="rounded-md border border-border p-8 text-center text-sm text-red-600">
          Failed to load customers. Please try again.
        </div>
      )}

      {/* Empty State */}
      {!isLoading && !error && customers.length === 0 && (
        <div className="rounded-md border border-border p-8 text-center text-sm text-muted-foreground">
          No customers found. {search && 'Try adjusting your search.'}
        </div>
      )}

      {/* Table */}
      {!isLoading && !error && customers.length > 0 && (
        <div className="rounded-md border border-border">
          <table className="w-full">
            <thead className="border-b border-border bg-muted/50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">
                  Name
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">
                  Email
                </th>
                <th className="px-4 py-3 text-right text-xs font-medium text-muted-foreground">
                  Total Purchases
                </th>
                <th className="px-4 py-3 text-center text-xs font-medium text-muted-foreground">
                  Status
                </th>
                <th className="px-4 py-3 text-right text-xs font-medium text-muted-foreground">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {customers.map((customer) => (
                <tr key={customer.id} className="hover:bg-muted/50">
                  <td className="px-4 py-3">
                    <Link href={`/customers/${customer.id}`} className="font-medium hover:underline">
                      {customer.name}
                    </Link>
                  </td>
                  <td className="px-4 py-3 text-sm text-muted-foreground">
                    {customer.email}
                  </td>
                  <td className="px-4 py-3 text-right text-sm">
                    ${(customer.total_purchases).toLocaleString()}
                  </td>
                  <td className="px-4 py-3 text-center">
                    <span className={`inline-flex rounded-full px-2 py-1 text-xs font-medium ${
                      customer.status === 'active' 
                        ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
                        : customer.status === 'inactive'
                        ? 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200'
                        : 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
                    }`}>
                      {customer.status}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-right">
                    <Link
                      href={`/customers/${customer.id}`}
                      className="text-sm text-muted-foreground hover:text-foreground"
                    >
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