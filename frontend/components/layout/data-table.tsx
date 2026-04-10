'use client'

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

type DataRow = {
  id: number
  header: string
  type: string
  status: string
  target: string
  limit: string
  reviewer: string
}

const statusConfig: Record<string, { dot: string; label: string }> = {
  'Done': { dot: 'bg-emerald-500', label: 'Done' },
  'In Process': { dot: 'bg-blue-500', label: 'In Process' },
}

function StatusBadge({ status }: { status: string }) {
  const config = statusConfig[status]
  if (!config) {
    return (
      <span className="inline-flex items-center gap-1.5 rounded-md border px-2 py-0.5 text-xs font-medium text-muted-foreground">
        {status}
      </span>
    )
  }
  return (
    <span className="inline-flex items-center gap-1.5 rounded-md border px-2 py-0.5 text-xs font-medium">
      <span className={`size-1.5 rounded-full ${config.dot}`} />
      {config.label}
    </span>
  )
}

export function DataTable({ data }: { data: DataRow[] }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Documents</CardTitle>
        <CardDescription>Manage your documents and files.</CardDescription>
      </CardHeader>
      <CardContent className="px-0">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="pl-6">Header</TableHead>
              <TableHead className="hidden md:table-cell">Type</TableHead>
              <TableHead>Status</TableHead>
              <TableHead className="hidden sm:table-cell">Target</TableHead>
              <TableHead className="hidden lg:table-cell">Limit</TableHead>
              <TableHead className="hidden xl:table-cell pr-6">Reviewer</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {data.map((row) => (
              <TableRow key={row.id}>
                <TableCell className="font-medium pl-6">{row.header}</TableCell>
                <TableCell className="hidden md:table-cell text-muted-foreground">{row.type}</TableCell>
                <TableCell><StatusBadge status={row.status} /></TableCell>
                <TableCell className="hidden sm:table-cell">{row.target}</TableCell>
                <TableCell className="hidden lg:table-cell">{row.limit}</TableCell>
                <TableCell className="hidden xl:table-cell pr-6 text-muted-foreground">{row.reviewer}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  )
}