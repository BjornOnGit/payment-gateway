import { NextResponse } from 'next/server'
import { API_BASE_URL } from '../../../../lib/config'

// Optional: support client credentials via env if backend requires a client
const CLIENT_ID = process.env.NEXT_PUBLIC_OAUTH_CLIENT_ID
const CLIENT_SECRET = process.env.NEXT_PUBLIC_OAUTH_CLIENT_SECRET

export async function POST(req: Request) {
  const contentType = req.headers.get('content-type') || ''
  let body: Record<string, string> = {}
  if (contentType.includes('application/json')) {
    body = await req.json()
  } else {
    const form = await req.formData()
    form.forEach((v, k) => { body[k] = String(v) })
  }

  // Username/password login using OAuth password grant (if supported by backend)
  const payload: Record<string, string> = {
    grant_type: 'password',
    // Send email as the username value for password grant
    username: body.email || '',
    password: body.password || '',
  }

  // Include client credentials if provided (some servers require a client even for password grant)
  const cid = body.client_id || CLIENT_ID
  const csec = body.client_secret || CLIENT_SECRET
  if (cid) payload.client_id = cid
  if (csec) payload.client_secret = csec

  const res = await fetch(`${API_BASE_URL}/oauth/token`, {
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
  response.cookies.set('access_token', token, {
    httpOnly: true,
    secure: true,
    sameSite: 'lax',
    path: '/',
    maxAge: 60 * 60, // 1h
  })
  return response
}
