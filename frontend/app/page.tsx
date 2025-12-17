export default function HomePage() {
  return (
    <div className="p-6 space-y-4">
      <h1 className="text-2xl font-semibold">Payment Gateway</h1>
      <p>Welcome. Please <a href="/login" className="underline">login</a> or go to the <a href="/dashboard" className="underline">dashboard</a>.</p>
    </div>
  )
}
