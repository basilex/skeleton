'use client'

import Link from 'next/link'

const invoices = [
  { id: '1', number: 'INV-2026-001', customer: 'Acme Corp', total: 12500, status: 'paid', date: '2026-04-01' },
  { id: '2', number: 'INV-2026-002', customer: 'Globex Inc', total: 8900, status: 'pending', date: '2026-04-03' },
  { id: '3', number: 'INV-2026-003', customer: 'Initech', total: 45200, status: 'overdue', date: '2026-03-15' },
]

export default function InvoicesPage() {
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
        <button className="rounded-md bg-secondary px-3 py-1.5 text-xs font-medium">
          All
        </button>
        <button className="rounded-md border border-border px-3 py-1.5 text-xs font-medium text-muted-foreground hover:bg-secondary">
          Draft
        </button>
        <button className="rounded-md border border-border px-3 py-1.5 text-xs font-medium text-muted-foreground hover:bg-secondary">
          Pending
        </button>
        <button className="rounded-md border border-border px-3 py-1.5 text-xs font-medium text-muted-foreground hover:bg-secondary">
          Paid
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
            {invoices.map((invoice) => (
              <tr key={invoice.id} className="hover:bg-muted/50">
                <td className="px-4 py-3">
                  <Link href={`/invoices/${invoice.id}`} className="font-medium hover:underline">
                    {invoice.number}
                  </Link>
                </td>
                <td className="px-4 py-3 text-sm text-muted-foreground">{invoice.customer}</td>
                <td className="px-4 py-3 text-right text-sm">${(invoice.total / 100).toLocaleString()}</td>
                <td className="px-4 py-3 text-center">
                  <span className={`inline-flex rounded-full px-2 py-1 text-xs font-medium ${
                    invoice.status === 'paid' ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200' :
                    invoice.status === 'pending' ? 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200' :
                    'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
                  }`}>
                    {invoice.status}
                  </span>
                </td>
                <td className="px-4 py-3 text-sm text-muted-foreground">{invoice.date}</td>
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
    </div>
  )
}