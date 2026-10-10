import { startAuthentication, startRegistration } from '@simplewebauthn/browser';

import axios from 'axios';
import {
  PasskeyLoginChallengeRequest,
  PasskeyLoginChallengeResponse,
  PasskeyLoginVerifyRequest,
  PasskeyLoginVerifyResponse,
  PasskeyRegisterChallengeRequest,
  PasskeyRegisterChallengeResponse,
  PasskeyRegisterVerifyRequest,
  PasskeyRegisterVerifyResponse,
} from '@sentinez/proto/sentinez/apps/iam/v1/iam';

import { API_BASE_PATH } from '@/lib/api/base';
import { toQuery } from '@/lib/api/pages';

// GET /iam/passkey/register/challenge
async function PasskeyRegisterChallenge(
  params: PasskeyRegisterChallengeRequest,
): Promise<PasskeyRegisterChallengeResponse> {
  const endpoint = `${API_BASE_PATH}/iam/passkey/register/challenge`;
  const resp = await axios.get(endpoint, {
    params: toQuery(PasskeyRegisterChallengeRequest.toJSON(params)),
  });
  return PasskeyRegisterChallengeResponse.fromJSON(resp.data ?? {});
}

// POST /iam/passkey/register/verify
async function PasskeyRegisterVerify(
  params: PasskeyRegisterVerifyRequest,
): Promise<PasskeyRegisterVerifyResponse> {
  const endpoint = `${API_BASE_PATH}/iam/passkey/register/verify`;
  const resp = await axios.post(endpoint, PasskeyRegisterVerifyRequest.toJSON(params));
  return PasskeyRegisterVerifyResponse.fromJSON(resp.data ?? {});
}

// GET /iam/passkey/login/challenge
async function PasskeyLoginChallenge(
  params: PasskeyLoginChallengeRequest,
): Promise<PasskeyLoginChallengeResponse> {
  const endpoint = `${API_BASE_PATH}/iam/passkey/login/challenge`;
  const resp = await axios.get(endpoint, {
    params: toQuery(PasskeyLoginChallengeRequest.toJSON(params)),
  });
  return PasskeyLoginChallengeResponse.fromJSON(resp.data ?? {});
}

// PUT /iam/passkey/login/verify
async function PasskeyLoginVerify(
  params: PasskeyLoginVerifyRequest,
): Promise<PasskeyLoginVerifyResponse> {
  const endpoint = `${API_BASE_PATH}/iam/passkey/login/verify`;
  const resp = await axios.put(endpoint, PasskeyLoginVerifyRequest.toJSON(params));
  return PasskeyLoginVerifyResponse.fromJSON(resp.data ?? {});
}

// bytes fields carry the WebAuthn credential as JSON
const credentialBytes = (credential: unknown) =>
  new TextEncoder().encode(JSON.stringify(credential));

export async function PasskeyRegister(
  params: PasskeyRegisterChallengeRequest,
): Promise<PasskeyRegisterVerifyResponse> {
  const start = await PasskeyRegisterChallenge(params);
  const options = start.options;
  if (options == undefined) {
    throw new Error('publicKey is null');
  }

  const credential = await startRegistration({
    optionsJSON: options['publicKey'],
  });

  return PasskeyRegisterVerify({
    sessionId: start.sessionId,
    credentialCreationResponse: credentialBytes(credential),
  });
}

export async function PasskeyLogin(
  params: PasskeyLoginChallengeRequest,
): Promise<PasskeyLoginVerifyResponse> {
  const start = await PasskeyLoginChallenge(params);
  const options = start.options;
  if (options == undefined) {
    throw new Error('publicKey is null');
  }

  const credential = await startAuthentication({
    optionsJSON: options['publicKey'],
  });

  return PasskeyLoginVerify({
    sessionId: start.sessionId,
    credentialAssertionData: credentialBytes(credential),
  });
}
