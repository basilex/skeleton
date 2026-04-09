'use client'

import { ReactNode } from 'react'
import { RequireAuth } from '@/lib/auth'
import { Sidebar } from '@/components/layout/sidebar'

export default function AppLayout({
  children,
}: {
  children: ReactNode
}) {
  return (
    <RequireAuth>
      <div className="flex h-screen bg-background">
        <Sidebar />
        <main className="flex-1 overflow-y-auto">
          {children}
        </main>
      </div>
    </RequireAuth>
  )
}