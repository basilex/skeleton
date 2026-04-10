'use client'

import { useQuery } from '@tanstack/react-query'
import { apiClient } from '../api/client'
import type { Invoice, PageResult } from '../../types/api'

export interface ListInvoicesParams {
  status?: string
  customer_id?: string
  cursor?: string
  limit?: number
}

export function useInvoices(params: ListInvoicesParams = {}) {
  return useQuery({
    queryKey: ['invoices', params],
    queryFn: async () => {
      const searchParams = new URLSearchParams()
      if (params.status) searchParams.set('status', params.status)
      if (params.customer_id) searchParams.set('customer_id', params.customer_id)
      if (params.cursor) searchParams.set('cursor', params.cursor)
      if (params.limit) searchParams.set('limit', String(params.limit))
      
      const query = searchParams.toString()
      const endpoint = query ? `/api/v1/invoices?${query}` : '/api/v1/invoices'
      
      return apiClient.get<PageResult<Invoice>>(endpoint)
    },
  })
}

export function useInvoice(id: string) {
  return useQuery({
    queryKey: ['invoice', id],
    queryFn: () => apiClient.get<Invoice>(`/api/v1/invoices/${id}`),
    enabled: !!id,
  })
}