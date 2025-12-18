"use client"
import { useQuery } from '@tanstack/react-query'
import { api } from '../../../lib/api'

type Settlement = {
  id: string
  merchant_account_id: string
  amount: number
  status: string
  created_at: string
}

export default function SettlementsPage() {
  const { data, isLoading, error } = useQuery({
    queryKey: ['settlements'],
    queryFn: async () => {
      const res = await api.get('/v1/settlements/list')
      return res.data.data as Settlement[]
    },
  })

  if (isLoading) return <div className="p-6">Loading...</div>
  if (error) return <div className="p-6">Error loading settlements</div>

  return (
    <div className="p-6">
      <h2 className="text-xl font-semibold mb-4">Settlements</h2>
      <div className="overflow-x-auto">
        <table className="min-w-full text-sm">
          <thead>
            <tr className="text-left border-b">
              <th className="py-2">ID</th>
              <th className="py-2">Amount</th>
              <th className="py-2">Status</th>
              <th className="py-2">Created</th>
            </tr>
          </thead>
          <tbody>
            {data?.map((s) => (
              <tr key={s.id} className="border-b">
                <td className="py-2">{s.id}</td>
                <td className="py-2">{s.amount}</td>
                <td className="py-2">{s.status}</td>
                <td className="py-2">{new Date(s.created_at).toLocaleString()}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
