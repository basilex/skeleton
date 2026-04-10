'use client'

import { ReactNode } from 'react'
import { RequireAuth } from '@/lib/auth'
import { SidebarProvider, SidebarInset } from '@/components/ui/sidebar'
import { AppSidebar } from '@/components/layout/app-sidebar'
import { SiteHeader } from '@/components/layout/site-header'

export default function AppLayout({ children }: { children: ReactNode }) {
  return (
    <RequireAuth>
      <SidebarProvider>
        <AppSidebar variant="inset" />
        <SidebarInset>
          <SiteHeader />
          {children}
        </SidebarInset>
      </SidebarProvider>
    </RequireAuth>
  )
}