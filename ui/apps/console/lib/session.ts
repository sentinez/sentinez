import { SENTINEZ_ACCESS_TOKEN_KEY, SENTINEZ_TENANT_KEY } from '@/lib/const';

const AUTH_PATH = '/auth';
const TENANT_PATH = '/tenant';
const CONSOLE_PATH = '/console';
const DOMAIN_PATH = '/console/domain';

// Expiry of a JWT in epoch milliseconds, or 0 when it cannot be read
function expiryOf(token: string): number {
  try {
    const payload = token.split('.')[1] ?? '';
    const json = atob(payload.replace(/-/g, '+').replace(/_/g, '/'));
    return Number(JSON.parse(json).exp) * 1000 || 0;
  } catch {
    return 0;
  }
}

/**
 * Whether the browser holds an access token that has not expired. An expired
 * token is dropped, so the login page stays reachable. Client-side only.
 */
export function isSignedIn(): boolean {
  const token = localStorage.getItem(SENTINEZ_ACCESS_TOKEN_KEY);
  if (!token) return false;
  if (expiryOf(token) > Date.now()) return true;

  localStorage.removeItem(SENTINEZ_ACCESS_TOKEN_KEY);
  return false;
}

/**
 * Where a signed-in user lands: the console once a tenant was picked,
 * otherwise the tenant picker. Client-side only.
 */
export function homePath(): string {
  return localStorage.getItem(SENTINEZ_TENANT_KEY) ? DOMAIN_PATH : TENANT_PATH;
}

function isUnder(pathname: string, base: string): boolean {
  return pathname === base || pathname.startsWith(`${base}/`);
}

/**
 * Where a visitor of pathname must go instead, or null when the page may be
 * shown: the auth pages are for guests, the tenant picker and the console
 * need a signed-in user, and the domain pages a tenant picked first.
 * Client-side only.
 */
export function redirectFor(pathname: string): string | null {
  const signedIn = isSignedIn();

  if (isUnder(pathname, AUTH_PATH)) return signedIn ? homePath() : null;

  const needsSignIn = isUnder(pathname, TENANT_PATH) || isUnder(pathname, CONSOLE_PATH);
  if (needsSignIn && !signedIn) return AUTH_PATH;

  if (isUnder(pathname, DOMAIN_PATH) && homePath() !== DOMAIN_PATH) return TENANT_PATH;
  return null;
}
