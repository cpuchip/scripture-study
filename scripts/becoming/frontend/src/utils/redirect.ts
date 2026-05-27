// Allowlist for post-login redirect targets. Mirrors isAllowedRedirect in
// internal/auth/oauth.go — keep the two in sync if the allowed hosts change.
//
// Without an allowlist, ?redirect=https://evil.com/phish would ride the
// post-login trust into an attacker-controlled page.
export function isAllowedRedirect(target: string): boolean {
  if (!target) return false
  if (target.startsWith('/') && !target.startsWith('//')) return true
  try {
    const url = new URL(target)
    if (url.protocol !== 'http:' && url.protocol !== 'https:') return false
    return url.hostname === 'ibeco.me' || url.hostname.endsWith('.ibeco.me')
  } catch {
    return false
  }
}
