'use client'
import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { api } from '../../../lib/api'
import { ArrowLeft } from 'lucide-react'
import Link from 'next/link'

export default function CreateTransactionPage() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState(false)
  const router = useRouter()

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    setError('')
    setSuccess(false)
    setLoading(true)

    const formData = new FormData(e.currentTarget)
    
    const payload = {
      amount: parseInt(formData.get('amount') as string) * 100, // Convert to cents
      currency: formData.get('currency') as string,
      user_id: formData.get('user_id') as string,
      merchant_id: formData.get('merchant_id') as string,
      metadata: {
        description: formData.get('description') as string,
      }
    }

    try {
      const res = await api.post('/v1/transactions', payload, {
        headers: {
          'Idempotency-Key': crypto.randomUUID(),
        }
      })

      if (res.status === 201) {
        setSuccess(true)
        setTimeout(() => router.push('/dashboard/transactions'), 2000)
      }
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to create transaction')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="p-6 max-w-2xl">
      <Link href="/dashboard" className="inline-flex items-center gap-2 text-sm text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100 mb-6">
        <ArrowLeft className="h-4 w-4" />
        Back to Dashboard
      </Link>

      <div className="bg-white dark:bg-slate-900 rounded-lg border border-slate-200 dark:border-slate-800 p-6">
        <h1 className="text-2xl font-bold text-slate-900 dark:text-slate-100 mb-2">Create Transaction</h1>
        <p className="text-slate-600 dark:text-slate-400 mb-6">Create a new payment transaction</p>

        {success && (
          <div className="mb-6 p-4 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg">
            <p className="text-green-800 dark:text-green-200 font-medium">Transaction created successfully! Redirecting...</p>
          </div>
        )}

        {error && (
          <div className="mb-6 p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
            <p className="text-red-800 dark:text-red-200 font-medium">{error}</p>
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
              Amount (NGN)
            </label>
            <input
              name="amount"
              type="number"
              step="0.01"
              placeholder="100.00"
              className="w-full border border-slate-300 dark:border-slate-700 rounded-lg px-3 py-2 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100"
              required
            />
            <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">Enter amount in dollars</p>
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
              Currency
            </label>
            <select
              name="currency"
              title="Currency"
              className="w-full border border-slate-300 dark:border-slate-700 rounded-lg px-3 py-2 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100"
              required
            >
              <option value="NGN">NGN</option>
              {/* <option value="EUR">EUR</option>
              <option value="GBP">GBP</option> */}
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
              User ID
            </label>
            <input
              name="user_id"
              type="text"
              placeholder="00000000-0000-0000-0000-000000000000"
              className="w-full border border-slate-300 dark:border-slate-700 rounded-lg px-3 py-2 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 font-mono text-sm"
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
              Merchant ID
            </label>
            <input
              name="merchant_id"
              type="text"
              placeholder="00000000-0000-0000-0000-000000000000"
              className="w-full border border-slate-300 dark:border-slate-700 rounded-lg px-3 py-2 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 font-mono text-sm"
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
              Description (optional)
            </label>
            <textarea
              name="description"
              placeholder="Transaction description"
              rows={3}
              className="w-full border border-slate-300 dark:border-slate-700 rounded-lg px-3 py-2 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100"
            />
          </div>

          <button
            type="submit"
            disabled={loading || success}
            className="w-full bg-slate-900 dark:bg-slate-100 text-white dark:text-slate-900 px-4 py-2.5 rounded-lg font-medium hover:bg-slate-800 dark:hover:bg-slate-200 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {loading ? 'Creating...' : 'Create Transaction'}
          </button>
        </form>
      </div>
    </div>
  )
}
