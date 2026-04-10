'use client'

import { useQuery } from '@tanstack/react-query'
import { apiClient } from '../api/client'
import type { Customer, PageResult } from '../../types/api'

export interface ListCustomersParams {
  status?: string
  search?: string
  tax_id?: string
  cursor?: string
  limit?: number
}

export function useCustomers(params: ListCustomersParams = {}) {
  return useQuery({
    queryKey: ['customers', params],
    queryFn: async () => {
      const searchParams = new URLSearchParams()
      if (params.status) searchParams.set('status', params.status)
      if (params.search) searchParams.set('search', params.search)
      if (params.tax_id) searchParams.set('tax_id', params.tax_id)
      if (params.cursor) searchParams.set('cursor', params.cursor)
      if (params.limit) searchParams.set('limit', String(params.limit))
      
      const query = searchParams.toString()
      const endpoint = query ? `/api/v1/customers?${query}` : '/api/v1/customers'
      
      return apiClient.get<PageResult<Customer>>(endpoint)
    },
  })
}

export function useCustomer(id: string) {
  return useQuery({
    queryKey: ['customer', id],
    queryFn: () => apiClient.get<Customer>(`/api/v1/customers/${id}`),
    enabled: !!id,
  })
}