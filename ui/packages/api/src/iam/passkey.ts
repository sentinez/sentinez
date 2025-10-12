import {
  PasskeyRegisterFinishRequest,
  PasskeyRegisterFinishResponse,
  PasskeyRegisterStartResponse,
} from '@sentinez/proto/sentinez/core/iam/v1/iam';

async function PasskeyRegisterStart(
  emailOrUsername: string,
): Promise<PasskeyRegisterStartResponse> {
  const resp = PasskeyRegisterStartResponse.fromJSON({
    event: undefined,
    sessionId: '',
  });

  return resp;
}

async function PasskeyRegisterFinish(
  params: PasskeyRegisterFinishRequest,
): Promise<PasskeyRegisterFinishResponse> {
  const resp = PasskeyRegisterFinishResponse.fromJSON({});

  return resp;
}
