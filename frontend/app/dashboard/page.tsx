'use client'
import { useQuery } from '@tanstack/react-query'
import { api } from '../../lib/api'
import { CreditCard, Banknote, TrendingUp, Activity } from 'lucide-react'

type Transaction = {
  id: string
  amount: number
  currency: string
  status: string
  created_at: string
}

type Settlement = {
  id: string
  amount: number
  status: string
}

export default function DashboardPage() {
  const { data: transactions, isLoading: loadingTx } = useQuery({
    queryKey: ['transactions'],
    queryFn: async () => {
      const res = await api.get('/v1/transactions/list?limit=100')
      return res.data.data as Transaction[]
    },
  })

  const { data: settlements, isLoading: loadingSt } = useQuery({
    queryKey: ['settlements'],
    queryFn: async () => {
      const res = await api.get('/v1/settlements/list?limit=100')
      return res.data.data as Settlement[]
    },
  })

  const totalTransactions = transactions?.length || 0
  const totalSettlements = settlements?.length || 0
  const currency = transactions?.[0]?.currency || 'NGN'
  const totalRevenue = transactions?.reduce((sum, t) => sum + (t.amount || 0), 0) || 0
  const successfulTxs = transactions?.filter(t => t.status === 'completed' || t.status === 'success').length || 0

  const currencySymbols: Record<string, string> = {
    USD: '$',
    EUR: '€',
    GBP: '£',
    NGN: '₦',
  }
  const currencySymbol = currencySymbols[currency] || currency

  const stats = [
    { label: 'Total Transactions', value: totalTransactions, icon: CreditCard, color: 'bg-blue-500' },
    { label: 'Total Settlements', value: totalSettlements, icon: Banknote, color: 'bg-green-500' },
    { label: 'Total Revenue', value: totalRevenue, icon: TrendingUp, color: 'bg-purple-500' },
    { label: 'Successful Txs', value: successfulTxs, icon: Activity, color: 'bg-orange-500' },
  ]

  return (
    <div className="p-6 space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-slate-900 dark:text-slate-100">Dashboard</h1>
        <p className="text-slate-600 dark:text-slate-400 mt-1">Overview of your payment gateway</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {stats.map((stat) => {
          const Icon = stat.icon
          const displayValue = stat.label === 'Total Revenue' 
            ? `${currencySymbol}${totalRevenue.toLocaleString()}`
            : stat.value
          return (
            <div key={stat.label} className="bg-white dark:bg-slate-900 rounded-lg border border-slate-200 dark:border-slate-800 p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-slate-600 dark:text-slate-400">{stat.label}</p>
                  <p className="text-2xl font-bold text-slate-900 dark:text-slate-100 mt-2">{displayValue}</p>
                </div>
                <div className={`${stat.color} p-3 rounded-lg`}>
                  <Icon className="h-6 w-6 text-white" />
                </div>
              </div>
            </div>
          )
        })}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-white dark:bg-slate-900 rounded-lg border border-slate-200 dark:border-slate-800 p-6">
          <h2 className="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-4">Recent Transactions</h2>
          <div className="space-y-3">
            {!loadingTx && transactions?.slice(0, 5).map((t) => (
              <div key={t.id} className="flex items-center justify-between py-2 border-b border-slate-100 dark:border-slate-800 last:border-0">
                <div>
                  <p className="text-sm font-medium text-slate-900 dark:text-slate-100">{t.id}</p>
                  <p className="text-xs text-slate-500 dark:text-slate-400">{new Date(t.created_at).toLocaleDateString()}</p>
                </div>
                <div className="text-right">
                  <p className="text-sm font-semibold text-slate-900 dark:text-slate-100">{currencySymbol}{t.amount.toLocaleString()}</p>
                  <span className={`text-xs px-2 py-0.5 rounded-full ${
                    t.status === 'completed' || t.status === 'success' ? 'bg-green-100 text-green-700' : 'bg-yellow-100 text-yellow-700'
                  }`}>
                    {t.status}
                  </span>
                </div>
              </div>
            ))}
            {loadingTx && <p className="text-sm text-slate-500 dark:text-slate-400">Loading transactions…</p>}
            {!loadingTx && !transactions?.length && <p className="text-sm text-slate-500 dark:text-slate-400">No transactions yet</p>}
          </div>
        </div>

        <div className="bg-white dark:bg-slate-900 rounded-lg border border-slate-200 dark:border-slate-800 p-6">
          <h2 className="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-4">Recent Settlements</h2>
          <div className="space-y-3">
            {!loadingSt && settlements?.slice(0, 5).map((s) => (
              <div key={s.id} className="flex items-center justify-between py-2 border-b border-slate-100 dark:border-slate-800 last:border-0">
                <div>
                  <p className="text-sm font-medium text-slate-900 dark:text-slate-100">{s.id}</p>
                </div>
                <div className="text-right">
                  <p className="text-sm font-semibold text-slate-900 dark:text-slate-100">{currencySymbol}{s.amount.toLocaleString()}</p>
                  <span className={`text-xs px-2 py-0.5 rounded-full ${
                    s.status === 'completed' ? 'bg-green-100 text-green-700' : 'bg-yellow-100 text-yellow-700'
                  }`}>
                    {s.status}
                  </span>
                </div>
              </div>
            ))}
            {loadingSt && <p className="text-sm text-slate-500 dark:text-slate-400">Loading settlements…</p>}
            {!loadingSt && !settlements?.length && <p className="text-sm text-slate-500 dark:text-slate-400">No settlements yet</p>}
          </div>
        </div>
      </div>
    </div>
  )
}
