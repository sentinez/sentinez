import axios from 'axios';
import {
  CreateAccountRequest,
  CreateAccountResponse,
  LoginRequest,
  LoginResponse,
} from '@sentinez/proto/sentinez/apps/iam/v1/iam';

import { API_BASE_PATH } from '@/lib/api/base';
import { USE_SAMPLE } from '@/lib/api/fallback';
import { sampleLogin } from '@/lib/faker/iam';

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
