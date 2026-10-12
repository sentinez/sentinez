import axios from 'axios';
import {
  CreateAccountRequest,
  CreateAccountResponse,
  LoginRequest,
  LoginResponse,
} from '@sentinez/proto/sentinez/apps/iam/v1/iam';

import { API_BASE_PATH } from '@/lib/api/base';

// PUT /iam/login
export async function Login(params: LoginRequest): Promise<LoginResponse> {
  const endpoint = `${API_BASE_PATH}/iam/login`;
  const resp = await axios.put(endpoint, LoginRequest.toJSON(params));
  return LoginResponse.fromJSON(resp.data ?? {});
}

// POST /iam/account
export async function CreateAccount(params: CreateAccountRequest): Promise<CreateAccountResponse> {
  const endpoint = `${API_BASE_PATH}/iam/account`;
  const resp = await axios.post(endpoint, CreateAccountRequest.toJSON(params));
  return CreateAccountResponse.fromJSON(resp.data ?? {});
}
