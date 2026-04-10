'use client'

import { ChartAreaInteractive } from '@/components/layout/chart-area-interactive'
import { DataTable } from '@/components/layout/data-table'
import { SectionCards } from '@/components/layout/section-cards'

import data from './data.json'

export default function DashboardPage() {
  return (
    <div className="flex flex-1 flex-col">
      <div className="@container/main flex flex-1 flex-col gap-2">
        <div className="flex flex-col gap-4 px-4 py-4 md:gap-6 md:px-6 md:py-6">
          <SectionCards />
          <ChartAreaInteractive />
          <DataTable data={data} />
        </div>
      </div>
    </div>
  )
}