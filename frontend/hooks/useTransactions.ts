"use client"
import { useQuery } from '@tanstack/react-query'
import { api } from '../lib/api'
import type { Transaction } from '../types/transaction'

export function useTransactions() {
  return useQuery({
    queryKey: ['transactions'],
    queryFn: async () => (await api.get('/transactions')).data as Transaction[],
  })
}
