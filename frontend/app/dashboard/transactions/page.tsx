"use client"
import { useQuery } from '@tanstack/react-query'
import { api } from '../../../lib/api'

type Transaction = {
  id: string
  amount: number
  currency: string
  status: string
  created_at: string
}

export default function TransactionsPage() {
  const { data, isLoading, error } = useQuery({
    queryKey: ['transactions'],
    queryFn: async () => {
      const res = await api.get('/v1/transactions/list')
      return res.data.data as Transaction[]
    },
  })

  if (isLoading) return <div className="p-6">Loading...</div>
  if (error) return <div className="p-6">Error loading transactions</div>

  return (
    <div className="p-6">
      <h2 className="text-xl font-semibold mb-4">Transactions</h2>
      <div className="overflow-x-auto">
        <table className="min-w-full text-sm">
          <thead>
            <tr className="text-left border-b">
              <th className="py-2">ID</th>
              <th className="py-2">Amount</th>
              <th className="py-2">Currency</th>
              <th className="py-2">Status</th>
              <th className="py-2">Created</th>
            </tr>
          </thead>
          <tbody>
            {data?.map((t) => (
              <tr key={t.id} className="border-b">
                <td className="py-2">{t.id}</td>
                <td className="py-2">{t.amount}</td>
                <td className="py-2">{t.currency}</td>
                <td className="py-2">{t.status}</td>
                <td className="py-2">{new Date(t.created_at).toLocaleString()}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
