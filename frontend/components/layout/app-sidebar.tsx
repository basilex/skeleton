'use client'

import { IconInnerShadowTop } from '@tabler/icons-react'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
} from '@/components/ui/sidebar'
import { NavMain } from './nav-main'
import { NavSecondary } from './nav-secondary'
import { NavDocuments } from './nav-documents'
import { NavUser } from './nav-user'

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  return (
    <Sidebar collapsible="offcanvas" {...props}>
      <SidebarHeader>
        <div className="flex items-center gap-2">
          <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-foreground text-background">
            <IconInnerShadowTop className="size-4" />
          </div>
          <div className="flex flex-col text-left">
            <span className="truncate text-sm font-semibold">Skeleton CRM</span>
            <span className="truncate text-xs text-muted-foreground">Customer Relations</span>
          </div>
        </div>
      </SidebarHeader>
      <SidebarContent>
        <NavMain />
        <NavSecondary />
        <NavDocuments />
      </SidebarContent>
      <SidebarFooter>
        <NavUser />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  )
}