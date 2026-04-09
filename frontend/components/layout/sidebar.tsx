'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { useAuth } from '@/lib/auth'
import { cn } from '@/lib/utils'

const navigation = [
  {
    name: 'Dashboard',
    href: '/dashboard',
    icon: '📊',
  },
  {
    name: 'Customers',
    href: '/customers',
    icon: '👥',
  },
  {
    name: 'Invoices',
    href: '/invoices',
    icon: '📄',
  },
  {
    name: 'Orders',
    href: '/orders',
    icon: '📦',
  },
  {
    name: 'Inventory',
    href: '/inventory',
    icon: '🏪',
  },
  {
    name: 'Products',
    href: '/products',
    icon: '🏷️',
  },
  {
    name: 'Accounting',
    href: '/accounting',
    icon: '💰',
  },
]

export function Sidebar() {
  const pathname = usePathname()
  const { user, logout } = useAuth()

  return (
    <div className="flex h-screen flex-col border-r border-border bg-background">
      {/* Header */}
      <div className="flex h-14 items-center border-b border-border px-4">
        <h1 className="text-lg font-semibold">Skeleton CRM</h1>
      </div>

      {/* Navigation */}
      <nav className="flex-1 space-y-1 overflow-y-auto p-2">
        {navigation.map((item) => {
          const isActive = pathname === item.href || pathname.startsWith(item.href + '/')
          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                'flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium',
                isActive
                  ? 'bg-secondary text-secondary-foreground'
                  : 'text-muted-foreground hover:bg-secondary/50 hover:text-foreground'
              )}
            >
              <span className="text-base">{item.icon}</span>
              {item.name}
            </Link>
          )
        })}
      </nav>

      {/* Footer */}
      <div className="border-t border-border p-4">
        <div className="flex items-center gap-2">
          <div className="flex-1 overflow-hidden">
            <p className="truncate text-sm font-medium">{user?.name}</p>
            <p className="truncate text-xs text-muted-foreground">{user?.email}</p>
          </div>
          <button
            onClick={logout}
            className="rounded-md border border-border bg-background px-3 py-1 text-xs font-medium hover:bg-secondary"
          >
            Sign out
          </button>
        </div>
      </div>
    </div>
  )
}