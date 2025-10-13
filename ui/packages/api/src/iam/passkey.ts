import {
  PasskeyLoginFinishRequest,
  PasskeyLoginFinishResponse,
  PasskeyLoginStartRequest,
  PasskeyLoginStartResponse,
  PasskeyRegisterFinishRequest,
  PasskeyRegisterFinishResponse,
  PasskeyRegisterStartRequest,
  PasskeyRegisterStartResponse,
} from '@sentinez/proto/sentinez/core/iam/v1/iam';

import axios from 'axios';

const API_BASE_PATH = process.env.SNTZ_BASE_PATH;

async function PasskeyRegisterStart(
  params: PasskeyRegisterStartRequest,
): Promise<PasskeyRegisterStartResponse> {
  try {
    const endpoint = `${API_BASE_PATH}/iam/passkey/register-start`;
    const resp = await axios.get<PasskeyRegisterStartResponse>(endpoint, {
      params,
    });

    return resp.data;
  } catch (error) {
    throw error;
  }
}

async function PasskeyRegisterFinish(
  params: PasskeyRegisterFinishRequest,
): Promise<PasskeyRegisterFinishResponse> {
  try {
    const endpoint = `${API_BASE_PATH}/iam/passkey/register-finish`;
    const body = PasskeyRegisterFinishRequest.toJSON(params);
    const resp = await axios.post<PasskeyRegisterFinishResponse>(endpoint, body);

    return resp.data;
  } catch (error) {
    throw error;
  }
}

async function PasskeyLoginStart(
  params: PasskeyLoginStartRequest,
): Promise<PasskeyLoginStartResponse> {
  try {
    const endpoint = `${API_BASE_PATH}/iam/passkey/login-start`;
    const resp = await axios.get<PasskeyLoginStartResponse>(endpoint, { params });

    return resp.data;
  } catch (error) {
    throw error;
  }
}

async function PasskeyLoginFinish(
  params: PasskeyLoginFinishRequest,
): Promise<PasskeyLoginFinishResponse> {
  try {
    const endpoint = `${API_BASE_PATH}/iam/passkey/login-finish`;
    const body = PasskeyLoginFinishRequest.toJSON(params);
    const resp = await axios.put<PasskeyLoginFinishResponse>(endpoint, body);

    return resp.data;
  } catch (error) {
    throw error;
  }
}
