import { cookies } from 'next/headers'
import { NextResponse } from 'next/server'
import { API_BASE_URL } from '../../../../lib/config'

async function proxy(req: Request, segments: string[]) {
  const token = cookies().get('access_token')?.value
  const url = `${API_BASE_URL}/${segments.join('/')}`

  const init: RequestInit = {
    method: req.method,
    headers: {
      'content-type': req.headers.get('content-type') || 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      // Forward idempotency
      ...(req.headers.get('Idempotency-Key') ? { 'Idempotency-Key': req.headers.get('Idempotency-Key')! } : {}),
    },
    body: ['GET', 'HEAD'].includes(req.method) ? undefined : await req.text(),
  }

  const resp = await fetch(url, init)
  const data = await resp.text()
  return new NextResponse(data, { status: resp.status, headers: resp.headers })
}

export async function GET(req: Request, { params }: { params: { path: string[] } }) {
  return proxy(req, params.path)
}
export async function POST(req: Request, { params }: { params: { path: string[] } }) {
  return proxy(req, params.path)
}
export async function PUT(req: Request, { params }: { params: { path: string[] } }) {
  return proxy(req, params.path)
}
export async function DELETE(req: Request, { params }: { params: { path: string[] } }) {
  return proxy(req, params.path)
}
