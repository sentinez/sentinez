import { PasskeyLoginChallengeRequest, PasskeyLoginVerifyResponse, PasskeyRegisterChallengeRequest, PasskeyRegisterVerifyResponse } from '@sentinez/proto/sentinez/modules/iam/v1/iam';
export declare function PasskeyRegister(params: PasskeyRegisterChallengeRequest): Promise<PasskeyRegisterVerifyResponse>;
export declare function PasskeyLogin(params: PasskeyLoginChallengeRequest): Promise<PasskeyLoginVerifyResponse>;
//# sourceMappingURL=passkey.d.ts.map