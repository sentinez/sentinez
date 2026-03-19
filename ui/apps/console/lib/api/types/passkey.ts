import { PublicKeyCredentialCreationOptionsJSON } from '@simplewebauthn/browser';

export interface PasskeyRegisterVerifyRequest {
  sessionId: string;
  credentialCreationResponse: string;
}

export interface PasskeyLoginVerifyRequest {
  sessionId: string;
  credentialAssertionData: string;
}
