'use client'

import Link from 'next/link'
import { IconFileDescription, IconFileText, IconFileSpreadsheet } from '@tabler/icons-react'
import {
  SidebarGroup,
  SidebarGroupLabel,
  SidebarGroupContent,
  SidebarMenu,
  SidebarMenuItem,
  SidebarMenuButton,
} from '@/components/ui/sidebar'

const documents = [
  { title: 'Proposals', href: '/documents/proposals', icon: IconFileDescription },
  { title: 'Contracts', href: '/documents/contracts', icon: IconFileText },
  { title: 'Reports', href: '/documents/reports', icon: IconFileSpreadsheet },
]

export function NavDocuments({
  items = documents,
}: {
  items?: { title: string; href: string; icon: React.ComponentType<{ className?: string }> }[]
}) {
  return (
    <SidebarGroup>
      <SidebarGroupLabel>Documents</SidebarGroupLabel>
      <SidebarGroupContent>
        <SidebarMenu>
          {items.map((item) => (
            <SidebarMenuItem key={item.title}>
              <SidebarMenuButton render={<Link href={item.href} />} tooltip={item.title}>
                <item.icon />
                <span>{item.title}</span>
              </SidebarMenuButton>
            </SidebarMenuItem>
          ))}
        </SidebarMenu>
      </SidebarGroupContent>
    </SidebarGroup>
  )
}