import axios from 'axios';
import {
  CreateAccountRequest,
  CreateAccountResponse,
  LoginRequest,
  LoginResponse,
} from '@sentinez/proto/sentinez/apps/iam/v1/iam';
import { User } from '@sentinez/proto/sentinez/apps/iam/v1/model';
import { ControlPlane } from '@sentinez/proto/sentinez/types/v1/known';

import { API_BASE_PATH } from '@/lib/api/base';
import { ALL_SAMPLE_USERS } from '@/lib/api/iam/user';

// Set NEXT_PUBLIC_USE_SAMPLE=true to sign in with a sample user when the
// backend is out of reach
const USE_SAMPLE = process.env.NEXT_PUBLIC_USE_SAMPLE === 'true';

// Lifetime of a sample access token
const SAMPLE_TOKEN_TTL_SECONDS = 24 * 60 * 60;

// PUT /iam/login
export async function Login(params: LoginRequest): Promise<LoginResponse> {
  const endpoint = `${API_BASE_PATH}/iam/login`;
  try {
    const resp = await axios.put(endpoint, LoginRequest.toJSON(params));
    return LoginResponse.fromJSON(resp.data ?? {});
  } catch (err: unknown) {
    // A refusal of the server (wrong password...) is thrown: only a backend
    // out of reach signs in with a sample user
    if (!USE_SAMPLE || !isOutOfReach(err)) throw err;
    console.warn('[iam] Login failed, using sample response:', (err as Error)?.message);
    return sampleLogin(params.emailOrUsername);
  }
}

// POST /iam/account
export async function CreateAccount(params: CreateAccountRequest): Promise<CreateAccountResponse> {
  const endpoint = `${API_BASE_PATH}/iam/account`;
  const resp = await axios.post(endpoint, CreateAccountRequest.toJSON(params));
  return CreateAccountResponse.fromJSON(resp.data ?? {});
}

// Statuses of the /api proxy (app/api/[...path]) or of the host when the
// apiserver cannot be reached
const UNREACHABLE_STATUSES = new Set([502, 503, 504]);

// Whether err tells of an apiserver out of reach rather than of its refusal
function isOutOfReach(err: unknown): boolean {
  if (axios.isCancel(err) || !axios.isAxiosError(err)) return false;
  const status = err.response?.status;
  return status === undefined || UNREACHABLE_STATUSES.has(status);
}

/**
 * Signs in as the sample user whose email or first name matches
 * emailOrUsername (any password), else as a new one named after it.
 */
function sampleLogin(emailOrUsername: string): LoginResponse {
  const key = emailOrUsername.trim().toLowerCase();
  const user =
    ALL_SAMPLE_USERS.find((u) => u.email === key || u.email.split('@')[0] === key) ??
    User.fromPartial({
      id: 'user_sample',
      fullName: emailOrUsername,
      email: key.includes('@') ? key : `${key}@example.com`,
      controlPlane: ControlPlane.CONTROL_PLANE_PORTAL,
    });
  return { user, accessToken: sampleToken(user.id) };
}

// An unsigned JWT: the console only reads its expiry
function sampleToken(sub: string): string {
  const b64 = (o: object) =>
    btoa(JSON.stringify(o)).replace(/=+$/, '').replace(/\+/g, '-').replace(/\//g, '_');
  const exp = Math.floor(Date.now() / 1000) + SAMPLE_TOKEN_TTL_SECONDS;
  return `${b64({ alg: 'none', typ: 'JWT' })}.${b64({ sub, exp })}.sample`;
}
