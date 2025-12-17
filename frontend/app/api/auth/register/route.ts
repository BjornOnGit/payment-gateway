import { NextResponse } from 'next/server'
import { API_BASE_URL } from '../../../../lib/config'

export async function POST(req: Request) {
  const contentType = req.headers.get('content-type') || ''
  let body: Record<string, string> = {}
  if (contentType.includes('application/json')) {
    body = await req.json()
  } else {
    const form = await req.formData()
    form.forEach((v, k) => {
      body[k] = String(v)
    })
  }

  const payload = {
    email: body.email || '',
    password: body.password || '',
  }

  const res = await fetch(`${API_BASE_URL}/auth/register`, {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify(payload),
  })

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
    return NextResponse.json(data || { message: 'Registration failed' }, { status: res.status })
  }

  const token = data.access_token || data.token || ''
  const response = NextResponse.json({ ok: true, user: data.user })
  if (token) {
    response.cookies.set('access_token', token, {
      httpOnly: true,
      secure: true,
      sameSite: 'lax',
      path: '/',
      maxAge: 60 * 60, // 1h
    })
  }
  return response
}
