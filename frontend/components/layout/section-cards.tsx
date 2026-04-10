'use client'

import {
  IconCoin,
  IconUsers,
  IconCreditCard,
  IconActivity,
  IconTrendingUp,
} from '@tabler/icons-react'
import { Card, CardHeader, CardDescription, CardAction, CardTitle, CardFooter } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'

const cards = [
  {
    title: 'Total Revenue',
    value: '$45,231.89',
    description: 'Monthly revenue',
    icon: IconCoin,
    trend: { value: '+12%', label: '+12% from last month' },
  },
  {
    title: 'Subscriptions',
    value: '+2,350',
    description: 'Active subscriptions',
    icon: IconUsers,
    trend: { value: '+2%', label: '+2% from last month' },
  },
  {
    title: 'Sales',
    value: '+12,234',
    description: 'Total sales this month',
    icon: IconCreditCard,
    trend: { value: '+8.1%', label: '+8.1% from last month' },
  },
  {
    title: 'Active Now',
    value: '+573',
    description: 'Currently active users',
    icon: IconActivity,
    trend: { value: '+201', label: '+201 since last hour' },
  },
]

export function SectionCards() {
  return (
    <div className="*:data-[slot=card]:shadow-none grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
      {cards.map((card) => (
        <Card key={card.title}>
          <CardHeader>
            <CardDescription>{card.title}</CardDescription>
            <CardAction>
              <Badge variant="outline" className="gap-1">
                <IconTrendingUp className="size-3" />
                {card.trend.value}
              </Badge>
            </CardAction>
            <CardTitle className="text-2xl font-semibold tabular-nums @[250px]/card:text-3xl">
              {card.value}
            </CardTitle>
          </CardHeader>
          <CardFooter className="flex-col items-start gap-1 text-sm">
            <div className="line-clamp-1 flex gap-2 text-muted-foreground">
              {card.trend.label}
            </div>
          </CardFooter>
        </Card>
      ))}
    </div>
  )
}