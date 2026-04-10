'use client'

import Link from 'next/link'
import { IconSettings, IconHelp } from '@tabler/icons-react'
import {
  SidebarGroup,
  SidebarGroupLabel,
  SidebarGroupContent,
  SidebarMenu,
  SidebarMenuItem,
  SidebarMenuButton,
} from '@/components/ui/sidebar'

const navSecondaryItems = [
  { title: 'Settings', href: '/settings', icon: IconSettings },
  { title: 'Get Help', href: 'https://help.skeletoncrm.com', icon: IconHelp },
]

export function NavSecondary({
  items = navSecondaryItems,
}: {
  items?: { title: string; href: string; icon: React.ComponentType<{ className?: string }> }[]
}) {
  return (
    <SidebarGroup>
      <SidebarGroupLabel>Support</SidebarGroupLabel>
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