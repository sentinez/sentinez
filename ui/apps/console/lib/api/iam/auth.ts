import axios from 'axios';
import { LoginRequest, LoginResponse } from '@sentinez/proto/sentinez/apps/iam/v1/iam';

const API_BASE_PATH = process.env.NEXT_PUBLIC_BASE_PATH || 'http://localhost:8080';

export async function Login(params: LoginRequest): Promise<LoginResponse> {
  try {
    const endpoint = `${API_BASE_PATH}/iam/login`;
    const resp = await axios.put<LoginResponse>(endpoint, params);
    return resp.data;
  } catch (error) {
    throw error;
  }
}
