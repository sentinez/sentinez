import { User } from '@sentinez/proto/sentinez/apps/iam/v1/model';
import { ControlPlane } from '@sentinez/proto/sentinez/types/v1/known';

import { SENTINEZ_ACCESS_TOKEN_KEY, SENTINEZ_TENANT_KEY, SENTINEZ_USER_KEY } from '@/lib/const';
import { StoredUser } from '@/lib/type/common';

const AUTH_PATH = '/auth';
const TENANT_PATH = '/tenant';
// The console and the settings
const APP_PATH = '/d';
const DOMAIN_PATH = '/d/console/domain';

const USER_CHANGED_EVENT = 'sentinez:user-changed';

/** Calls onChange whenever storeUser replaces the stored user. Client-side only. */
export function subscribeUser(onChange: () => void) {
  window.addEventListener(USER_CHANGED_EVENT, onChange);
  return () => window.removeEventListener(USER_CHANGED_EVENT, onChange);
}

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
 * Keeps the user of a login response in localStorage. `email` is what was
 * typed in the form, used when the response carries no user. Client-side only.
 */
export function storeUser(user: User | undefined, email = '') {
  const local: StoredUser = {
    id: user?.id ?? '',
    name: user?.fullName || email,
    email: user?.email || email,
    controlPlane: user?.controlPlane ?? ControlPlane.CONTROL_PLANE_UNSPECIFIED,
    createdAt: user?.metadata?.createdAt?.toISOString() ?? '',
    updatedAt: user?.metadata?.updatedAt?.toISOString() ?? '',
    fullName: user?.fullName ?? ''
  };
  localStorage.setItem(SENTINEZ_USER_KEY, JSON.stringify(local));
  window.dispatchEvent(new Event(USER_CHANGED_EVENT));
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
 * shown: the auth pages are for guests, the tenant picker, the console and
 * the settings need a signed-in user, and the domain pages a tenant picked first.
 * Client-side only.
 */
export function redirectFor(pathname: string): string | null {
  const signedIn = isSignedIn();

  if (isUnder(pathname, AUTH_PATH)) return signedIn ? homePath() : null;

  const needsSignIn = isUnder(pathname, TENANT_PATH) || isUnder(pathname, APP_PATH);
  if (needsSignIn && !signedIn) return AUTH_PATH;

  if (isUnder(pathname, DOMAIN_PATH) && homePath() !== DOMAIN_PATH) return TENANT_PATH;
  return null;
}
