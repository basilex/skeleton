'use client'

import Link from 'next/link'
import { useState } from 'react'
import { useInvoices } from '@/lib/query'

export default function InvoicesPage() {
  const [statusFilter, setStatusFilter] = useState<string | null>(null)
  const { data, isLoading, error } = useInvoices({ 
    status: statusFilter || undefined, 
    limit: 100 
  })

  const invoices = data?.items ?? []

  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Invoices</h1>
          <p className="text-sm text-muted-foreground">
            Manage customer invoices
          </p>
        </div>
        <Link
          href="/invoices/new"
          className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
        >
          + Create Invoice
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
          onClick={() => setStatusFilter('draft')}
          className={`rounded-md px-3 py-1.5 text-xs font-medium ${
            statusFilter === 'draft' 
              ? 'bg-secondary' 
              : 'border border-border text-muted-foreground hover:bg-secondary'
          }`}>
          Draft
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
          onClick={() => setStatusFilter('paid')}
          className={`rounded-md px-3 py-1.5 text-xs font-medium ${
            statusFilter === 'paid' 
              ? 'bg-secondary' 
              : 'border border-border text-muted-foreground hover:bg-secondary'
          }`}>
          Paid
        </button>
      </div>

      {/* Loading State */}
      {isLoading && (
        <div className="rounded-md border border-border p-8 text-center text-sm text-muted-foreground">
          Loading invoices...
        </div>
      )}

      {/* Error State */}
      {error && (
        <div className="rounded-md border border-border p-8 text-center text-sm text-red-600">
          Failed to load invoices. Please try again.
        </div>
      )}

      {/* Empty State */}
      {!isLoading && !error && invoices.length === 0 && (
        <div className="rounded-md border border-border p-8 text-center text-sm text-muted-foreground">
          No invoices found.
        </div>
      )}

      {/* Table */}
      {!isLoading && !error && invoices.length > 0 && (
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
              {invoices.map((invoice) => (
                <tr key={invoice.id} className="hover:bg-muted/50">
                  <td className="px-4 py-3">
                    <Link href={`/invoices/${invoice.id}`} className="font-medium hover:underline">
                      {invoice.number}
                    </Link>
                  </td>
                  <td className="px-4 py-3 text-sm text-muted-foreground">{invoice.customer_name}</td>
                  <td className="px-4 py-3 text-right text-sm">${invoice.total.toLocaleString()}</td>
                  <td className="px-4 py-3 text-center">
                    <span className={`inline-flex rounded-full px-2 py-1 text-xs font-medium capitalize ${
                      invoice.status === 'paid' ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200' :
                      invoice.status === 'pending' ? 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200' :
                      invoice.status === 'overdue' ? 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200' :
                      'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200'
                    }`}>
                      {invoice.status}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-sm text-muted-foreground">
                    {new Date(invoice.issue_date).toLocaleDateString()}
                  </td>
                  <td className="px-4 py-3 text-right">
                    <Link href={`/invoices/${invoice.id}`} className="text-sm text-muted-foreground hover:text-foreground">
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