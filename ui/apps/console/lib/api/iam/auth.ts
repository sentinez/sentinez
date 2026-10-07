import axios from 'axios';
import { LoginRequest, LoginResponse } from '@sentinez/proto/sentinez/apps/iam/v1/iam';

import { API_BASE_PATH } from '@/lib/api/base';

export async function Login(params: LoginRequest): Promise<LoginResponse> {
  try {
    const endpoint = `${API_BASE_PATH}/iam/login`;
    const resp = await axios.put<LoginResponse>(endpoint, params);
    return resp.data;
  } catch (error) {
    throw error;
  }
}
