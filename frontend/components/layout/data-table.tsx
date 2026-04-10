'use client'

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
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

function StatusBadge({ status }: { status: string }) {
  if (status === 'Done') {
    return <Badge variant="default" className="bg-emerald-500/15 text-emerald-500 border-emerald-500/20 hover:bg-emerald-500/25">{status}</Badge>
  }
  if (status === 'In Process') {
    return <Badge variant="default" className="bg-blue-500/15 text-blue-500 border-blue-500/20 hover:bg-blue-500/25">{status}</Badge>
  }
  return <Badge variant="secondary">{status}</Badge>
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