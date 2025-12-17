# Payment Gateway Frontend

Tech: Next.js (App Router), TypeScript, Tailwind CSS, React Query, Zustand, Axios.

## Setup

```bash
cd frontend
npm install
npm run dev
```

Set env:

```
NEXT_PUBLIC_API_BASE_URL=https://api.payment-gateway.rayzortestr.com
# Optional if backend requires a client for password grant
NEXT_PUBLIC_OAUTH_CLIENT_ID=
NEXT_PUBLIC_OAUTH_CLIENT_SECRET=
```

## Auth Flow
- Login posts to `/api/auth/login` with username/password. The route requests a token from `${NEXT_PUBLIC_API_BASE_URL}/oauth/token` using `grant_type=password` (optionally includes `client_id`/`client_secret` if required).
- Token is stored as httpOnly cookie `access_token`.
- Client requests go through `/api/proxy/*` which injects the `Authorization: Bearer` header server-side.

## Deploy (Vercel)
- Import the `frontend` folder as the project
- Env: `NEXT_PUBLIC_API_BASE_URL` set to your API domain
- Build command: `npm run build`
- Output: `.vercel/output` (default Next)
