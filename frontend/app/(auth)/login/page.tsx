export const dynamic = 'force-static'

export default function LoginPage() {
  return (
    <div className="mx-auto max-w-md p-6">
      <h1 className="text-2xl font-semibold mb-4">Login</h1>
      <form action="/api/auth/login" method="post" className="space-y-3">
        <div>
          <label className="block text-sm">Email</label>
          <input name="email" type="email" className="w-full border rounded px-3 py-2" required />
        </div>
        <div>
          <label className="block text-sm">Password</label>
          <input name="password" type="password" className="w-full border rounded px-3 py-2" required />
        </div>
        {/* Optional client credentials if your backend requires them; remove if not needed */}
        <details className="text-sm">
          <summary className="cursor-pointer">Advanced (client credentials)</summary>
          <div className="mt-2 space-y-2">
            <input name="client_id" placeholder="client_id (optional)" className="w-full border rounded px-3 py-2" />
            <input name="client_secret" placeholder="client_secret (optional)" type="password" className="w-full border rounded px-3 py-2" />
          </div>
        </details>
        <button className="bg-black text-white px-4 py-2 rounded">Login</button>
      </form>
    </div>
  )
}
