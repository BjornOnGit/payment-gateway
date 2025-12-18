import { NextResponse } from 'next/server'
import { API_BASE_URL } from '../../../../lib/config'

export async function POST(req: Request) {
  const contentType = req.headers.get('content-type') || ''
  let body: Record<string, string> = {}
  if (contentType.includes('application/json')) {
    body = await req.json()
  } else {
    const form = await req.formData()
    form.forEach((v, k) => { body[k] = String(v) })
  }

  const payload = {
    email: body.email || '',
    password: body.password || '',
  }

  const res = await fetch(`${API_BASE_URL}/auth/login`, {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify(payload),
  })

  // Some backends return plain text on error (e.g., "unsupported grant type")
  let data: any = null
  const contentTypeRes = res.headers.get('content-type') || ''
  if (contentTypeRes.includes('application/json')) {
    data = await res.json()
  } else {
    const text = await res.text()
    try {
      data = JSON.parse(text)
    } catch {
      data = { message: text }
    }
  }

  if (!res.ok) {
    return NextResponse.json(data || { message: 'Login failed' }, { status: res.status })
  }

  const token = data.access_token || data.token || ''
  const response = NextResponse.json({ ok: true })
  const isProduction = process.env.NODE_ENV === 'production'
  response.cookies.set('access_token', token, {
    httpOnly: true,
    secure: isProduction,
    sameSite: 'lax',
    path: '/',
    maxAge: 60 * 60, // 1h
  })
  return response
}
