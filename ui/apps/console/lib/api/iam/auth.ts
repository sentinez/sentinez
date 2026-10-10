import axios from 'axios';
import { LoginRequest, LoginResponse } from '@sentinez/proto/sentinez/apps/iam/v1/iam';

import { API_BASE_PATH } from '@/lib/api/base';

// PUT /iam/login
export async function Login(params: LoginRequest): Promise<LoginResponse> {
  const endpoint = `${API_BASE_PATH}/iam/login`;
  const resp = await axios.put(endpoint, LoginRequest.toJSON(params));
  return LoginResponse.fromJSON(resp.data ?? {});
}
