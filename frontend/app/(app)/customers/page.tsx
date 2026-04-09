'use client'

import Link from 'next/link'
import { useState } from 'react'

// Mock data - will be replaced with API call
const customers = [
  { id: '1', name: 'Acme Corporation', email: 'contact@acme.com', total: 125000, status: 'active' },
  { id: '2', name: 'Globex Inc', email: 'info@globex.com', total: 89500, status: 'active' },
  { id: '3', name: 'Initech', email: 'hello@initech.com', total: 45200, status: 'inactive' },
]

export default function CustomersPage() {
  const [search, setSearch] = useState('')

  const filtered = customers.filter(c => 
    c.name.toLowerCase().includes(search.toLowerCase()) ||
    c.email.toLowerCase().includes(search.toLowerCase())
  )

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

      {/* Table */}
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
            {filtered.map((customer) => (
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
                  ${(customer.total / 100).toLocaleString()}
                </td>
                <td className="px-4 py-3 text-center">
                  <span className={`inline-flex rounded-full px-2 py-1 text-xs font-medium ${
                    customer.status === 'active' 
                      ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
                      : 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200'
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
    </div>
  )
}