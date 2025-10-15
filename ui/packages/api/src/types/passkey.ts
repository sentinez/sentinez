import { PublicKeyCredentialCreationOptionsJSON } from '@simplewebauthn/browser';

export interface PasskeyOption {
  optionsJSON: PublicKeyCredentialCreationOptionsJSON;
  useAutoRegister?: boolean;
}

export interface PasskeyRegisterFinishRequest {
  sessionId: string;
  credentialCreationResponse: string;
}
