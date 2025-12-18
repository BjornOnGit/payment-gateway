'use client'
import type { ReactNode } from 'react'
import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { LayoutDashboard, CreditCard, Banknote, Plus, ChevronLeft, ChevronRight } from 'lucide-react'
import { useState } from 'react'

function Sidebar({ collapsed, onToggle }: { collapsed: boolean; onToggle: () => void }) {
  const pathname = usePathname()
  
  const links = [
    { href: '/dashboard', label: 'Overview', icon: LayoutDashboard },
    { href: '/dashboard/transactions', label: 'Transactions', icon: CreditCard },
    { href: '/dashboard/settlements', label: 'Settlements', icon: Banknote },
    { href: '/dashboard/create-transaction', label: 'New Transaction', icon: Plus },
  ]

  return (
    <aside className={`border-r bg-slate-50 dark:bg-slate-900 ${collapsed ? 'w-16' : 'w-64'} transition-[width] duration-200`}
      style={{ minHeight: '100vh' }}
    >
      <div className={`p-4 ${collapsed ? 'px-2' : 'px-4'}`}>
        <div className="flex items-center justify-between mb-4">
          <h2 className={`font-bold text-slate-900 dark:text-slate-100 ${collapsed ? 'text-base' : 'text-xl'}`}>{collapsed ? 'PG' : 'Payment Gateway'}</h2>
          <button
            onClick={onToggle}
            className="inline-flex items-center justify-center rounded-md border border-slate-200 dark:border-slate-800 hover:bg-slate-100 dark:hover:bg-slate-800 h-8 w-8"
            aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
            title={collapsed ? 'Expand' : 'Collapse'}
          >
            {collapsed ? <ChevronRight className="h-4 w-4" /> : <ChevronLeft className="h-4 w-4" />}
          </button>
        </div>

        <nav className="space-y-1 mt-2">
          {links.map((link) => {
            const Icon = link.icon
            const isActive = pathname === link.href
            return (
              <Link
                key={link.href}
                href={link.href}
                className={`flex items-center ${collapsed ? 'gap-0 justify-center' : 'gap-3'} rounded-lg px-3 py-2.5 text-sm font-medium transition-colors ${
                  isActive
                    ? 'bg-slate-900 text-white dark:bg-slate-100 dark:text-slate-900'
                    : 'text-slate-700 hover:bg-slate-200 dark:text-slate-300 dark:hover:bg-slate-800'
                }`}
              >
                <Icon className="h-5 w-5" />
                {!collapsed && <span>{link.label}</span>}
              </Link>
            )
          })}
        </nav>
      </div>
    </aside>
  )
}

export default function DashboardLayout({ children }: { children: ReactNode }) {
  const [collapsed, setCollapsed] = useState(false)
  return (
    <div className="min-h-screen flex">
      <Sidebar collapsed={collapsed} onToggle={() => setCollapsed((s) => !s)} />
      <main className="flex-1 bg-white dark:bg-slate-950">
        {children}
      </main>
    </div>
  )
}
