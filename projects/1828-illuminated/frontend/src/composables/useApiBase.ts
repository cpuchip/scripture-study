// Resolves the i1828 backend API base URL.
//
//   - Production / inside the compose stack: `/api` (relative — nginx
//     reverse-proxies same-origin to the backend container).
//   - Local `npm run dev` against the compose stack on host 8083: pulled
//     from `.env.development` as `http://localhost:8083/api`.
//   - Any override at build time via the VITE_API_BASE_URL env var wins.
//
// The inline fallback ensures a missing/empty env still produces a working
// production build — we never hard-code `localhost` for the deployed bundle.

const RAW = (import.meta.env.VITE_API_BASE_URL ?? '').trim()

/** Trailing-slash-trimmed base URL for the i1828 backend. */
export const API_BASE: string = (RAW || '/api').replace(/\/+$/, '')

/** Join a path onto API_BASE, normalizing leading slashes. */
export function apiUrl(path: string): string {
  if (!path) return API_BASE
  const p = path.startsWith('/') ? path : `/${path}`
  return `${API_BASE}${p}`
}
