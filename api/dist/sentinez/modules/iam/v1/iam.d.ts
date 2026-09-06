import { BinaryReader, BinaryWriter } from "@bufbuild/protobuf/wire";
import { Context } from "../../../types/v1/known";
import { Pages } from "../../../types/v1/model";
import { AccountResponse, User } from "./model";
export declare const protobufPackage = "sentinez.modules.iam.v1";
export interface PasskeyLoginVerifyRequest {
    sessionId: string;
    credentialAssertionData: Uint8Array;
}
export interface PasskeyLoginVerifyResponse {
    user?: User | undefined;
    accessToken: string;
}
export interface PasskeyLoginChallengeRequest {
    emailOrUsername: string;
}
export interface PasskeyLoginChallengeResponse {
    options?: {
        [key: string]: any;
    } | undefined;
    sessionId: string;
}
export interface PasskeyRegisterVerifyRequest {
    sessionId: string;
    credentialCreationResponse: Uint8Array;
}
export interface PasskeyRegisterVerifyResponse {
}
export interface PasskeyRegisterChallengeRequest {
    emailOrUsername: string;
}
export interface PasskeyRegisterChallengeResponse {
    options?: {
        [key: string]: any;
    } | undefined;
    sessionId: string;
}
export interface LoginRequest {
    emailOrUsername: string;
    password: string;
}
export interface LoginResponse {
    user?: User | undefined;
    accessToken: string;
}
export interface CreateAccountRequest {
    username: string;
    fullName: string;
    password: string;
    email: string;
    phoneNumber: string;
}
export interface CreateAccountResponse {
    accountId: string;
}
export interface CreateUserRequest {
    fullName: string;
    email: string;
    phoneNumber: string;
}
export interface CreateUserResponse {
    userId: string;
}
export interface StatusRequest {
}
export interface StatusResponse {
    msg: string;
    context?: Context | undefined;
}
export interface UpdateUserRequest {
    id: string;
    fullName: string;
    email: string;
    phoneNumber: string;
}
export interface UpdateUserResponse {
    user?: User | undefined;
}
export interface GetUserRequest {
    id: string;
    email: string;
}
export interface GetUserResponse {
    user?: User | undefined;
}
export interface ListUsersRequest {
    page?: Pages | undefined;
    ids: string[];
    phoneNumbers: string[];
    emails: string[];
}
export interface ListUsersResponse {
    users: User[];
    total: number;
}
export interface ListAccountsRequest {
    page?: Pages | undefined;
    ids: string[];
    userIds: string[];
    usernames: string[];
    emails: string[];
}
export interface ListAccountsResponse {
    accounts: AccountResponse[];
    total: number;
}
export interface DeleteUserRequest {
    id: string;
}
export interface DeleteUserResponse {
}
export declare const PasskeyLoginVerifyRequest: MessageFns<PasskeyLoginVerifyRequest>;
export declare const PasskeyLoginVerifyResponse: MessageFns<PasskeyLoginVerifyResponse>;
export declare const PasskeyLoginChallengeRequest: MessageFns<PasskeyLoginChallengeRequest>;
export declare const PasskeyLoginChallengeResponse: MessageFns<PasskeyLoginChallengeResponse>;
export declare const PasskeyRegisterVerifyRequest: MessageFns<PasskeyRegisterVerifyRequest>;
export declare const PasskeyRegisterVerifyResponse: MessageFns<PasskeyRegisterVerifyResponse>;
export declare const PasskeyRegisterChallengeRequest: MessageFns<PasskeyRegisterChallengeRequest>;
export declare const PasskeyRegisterChallengeResponse: MessageFns<PasskeyRegisterChallengeResponse>;
export declare const LoginRequest: MessageFns<LoginRequest>;
export declare const LoginResponse: MessageFns<LoginResponse>;
export declare const CreateAccountRequest: MessageFns<CreateAccountRequest>;
export declare const CreateAccountResponse: MessageFns<CreateAccountResponse>;
export declare const CreateUserRequest: MessageFns<CreateUserRequest>;
export declare const CreateUserResponse: MessageFns<CreateUserResponse>;
export declare const StatusRequest: MessageFns<StatusRequest>;
export declare const StatusResponse: MessageFns<StatusResponse>;
export declare const UpdateUserRequest: MessageFns<UpdateUserRequest>;
export declare const UpdateUserResponse: MessageFns<UpdateUserResponse>;
export declare const GetUserRequest: MessageFns<GetUserRequest>;
export declare const GetUserResponse: MessageFns<GetUserResponse>;
export declare const ListUsersRequest: MessageFns<ListUsersRequest>;
export declare const ListUsersResponse: MessageFns<ListUsersResponse>;
export declare const ListAccountsRequest: MessageFns<ListAccountsRequest>;
export declare const ListAccountsResponse: MessageFns<ListAccountsResponse>;
export declare const DeleteUserRequest: MessageFns<DeleteUserRequest>;
export declare const DeleteUserResponse: MessageFns<DeleteUserResponse>;
type Builtin = Date | Function | Uint8Array | string | number | boolean | undefined;
export type DeepPartial<T> = T extends Builtin ? T : T extends globalThis.Array<infer U> ? globalThis.Array<DeepPartial<U>> : T extends ReadonlyArray<infer U> ? ReadonlyArray<DeepPartial<U>> : T extends {} ? {
    [K in keyof T]?: DeepPartial<T[K]>;
} : Partial<T>;
type KeysOfUnion<T> = T extends T ? keyof T : never;
export type Exact<P, I extends P> = P extends Builtin ? P : P & {
    [K in keyof P]: Exact<P[K], I[K]>;
} & {
    [K in Exclude<keyof I, KeysOfUnion<P>>]: never;
};
export interface MessageFns<T> {
    encode(message: T, writer?: BinaryWriter): BinaryWriter;
    decode(input: BinaryReader | Uint8Array, length?: number): T;
    fromJSON(object: any): T;
    toJSON(message: T): unknown;
    create<I extends Exact<DeepPartial<T>, I>>(base?: I): T;
    fromPartial<I extends Exact<DeepPartial<T>, I>>(object: I): T;
}
export {};
