'use client'

import { useQuery } from '@tanstack/react-query'
import { apiClient } from '../api/client'
import type { Order, PageResult } from '../../types/api'

export interface ListOrdersParams {
  status?: string
  customer_id?: string
  cursor?: string
  limit?: number
}

export function useOrders(params: ListOrdersParams = {}) {
  return useQuery({
    queryKey: ['orders', params],
    queryFn: async () => {
      const searchParams = new URLSearchParams()
      if (params.status) searchParams.set('status', params.status)
      if (params.customer_id) searchParams.set('customer_id', params.customer_id)
      if (params.cursor) searchParams.set('cursor', params.cursor)
      if (params.limit) searchParams.set('limit', String(params.limit))
      
      const query = searchParams.toString()
      const endpoint = query ? `/api/v1/orders?${query}` : '/api/v1/orders'
      
      return apiClient.get<PageResult<Order>>(endpoint)
    },
  })
}

export function useOrder(id: string) {
  return useQuery({
    queryKey: ['order', id],
    queryFn: () => apiClient.get<Order>(`/api/v1/orders/${id}`),
    enabled: !!id,
  })
}