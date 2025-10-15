import {
  PasskeyLoginFinishRequest,
  PasskeyLoginFinishResponse,
  PasskeyLoginStartRequest,
  PasskeyLoginStartResponse,
  PasskeyRegisterFinishResponse,
  PasskeyRegisterStartRequest,
  PasskeyRegisterStartResponse,
} from '@sentinez/proto/sentinez/core/iam/v1/iam';
import { startRegistration } from '@simplewebauthn/browser';

import axios from 'axios';
import { PasskeyOption, PasskeyRegisterFinishRequest } from '@sentinez/api/types/passkey';

const API_BASE_PATH = process.env.SNTZ_BASE_PATH || 'http://localhost:8080';

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

    const resp = await axios.post<PasskeyRegisterFinishResponse>(endpoint, params);

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

export async function PasskeyRegister(
  params: PasskeyRegisterStartRequest,
): Promise<PasskeyRegisterFinishResponse> {
  try {
    const start = await PasskeyRegisterStart(params);

    const options = start.event ? start.event : undefined;
    if (options == undefined) {
      throw new Error('publicKey is null');
    }

    const credential = await startRegistration({
      optionsJSON: options['publicKey'],
    });

    const encoded = btoa(JSON.stringify(credential));

    const finish = await PasskeyRegisterFinish({
      sessionId: start.sessionId,
      credentialCreationResponse: encoded,
    });

    return finish;
  } catch (error) {
    throw error;
  }
}
