'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import {
  IconDashboard,
  IconUsers,
  IconFileInvoice,
  IconPackage,
  IconBuildingWarehouse,
  IconBox,
  IconCoin,
  IconChevronRight,
} from '@tabler/icons-react'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible'
import {
  SidebarGroup,
  SidebarGroupLabel,
  SidebarGroupContent,
  SidebarMenu,
  SidebarMenuItem,
  SidebarMenuButton,
  SidebarMenuSub,
  SidebarMenuSubButton,
  SidebarMenuSubItem,
} from '@/components/ui/sidebar'

interface NavItem {
  title: string
  href: string
  icon: React.ComponentType<{ className?: string }>
  subItems?: { title: string; href: string }[]
}

const navMainItems: NavItem[] = [
  { title: 'Dashboard', href: '/dashboard', icon: IconDashboard },
  { title: 'Customers', href: '/customers', icon: IconUsers },
  { title: 'Invoices', href: '/invoices', icon: IconFileInvoice },
  { title: 'Orders', href: '/orders', icon: IconPackage },
  { title: 'Inventory', href: '/inventory', icon: IconBuildingWarehouse },
  { title: 'Products', href: '/products', icon: IconBox },
  { title: 'Accounting', href: '/accounting', icon: IconCoin },
]

export function NavMain({ items = navMainItems }: { items?: NavItem[] }) {
  const pathname = usePathname()

  return (
    <SidebarGroup>
      <SidebarGroupLabel>Platform</SidebarGroupLabel>
      <SidebarGroupContent>
        <SidebarMenu>
          {items.map((item) => {
            const isActive =
              pathname === item.href || pathname.startsWith(item.href + '/')

            if (item.subItems?.length) {
              return (
                <Collapsible key={item.href} defaultOpen={isActive} className="group/collapsible">
                  <SidebarMenuItem>
                    <CollapsibleTrigger
                      render={<SidebarMenuButton tooltip={item.title} isActive={isActive} />}
                    >
                      <item.icon />
                      <span>{item.title}</span>
                      <IconChevronRight className="ml-auto transition-transform duration-200 group-data-[panel-open]/collapsible:rotate-90" />
                    </CollapsibleTrigger>
                    <CollapsibleContent>
                      <SidebarMenuSub>
                        {item.subItems.map((sub) => (
                          <SidebarMenuSubItem key={sub.href}>
                            <SidebarMenuSubButton
                              render={<Link href={sub.href} />}
                              isActive={pathname === sub.href}
                            >
                              <span>{sub.title}</span>
                            </SidebarMenuSubButton>
                          </SidebarMenuSubItem>
                        ))}
                      </SidebarMenuSub>
                    </CollapsibleContent>
                  </SidebarMenuItem>
                </Collapsible>
              )
            }

            return (
              <SidebarMenuItem key={item.href}>
                <SidebarMenuButton
                  render={<Link href={item.href} />}
                  isActive={isActive}
                  tooltip={item.title}
                >
                  <item.icon />
                  <span>{item.title}</span>
                </SidebarMenuButton>
              </SidebarMenuItem>
            )
          })}
        </SidebarMenu>
      </SidebarGroupContent>
    </SidebarGroup>
  )
}