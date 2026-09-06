import { BinaryReader, BinaryWriter } from "@bufbuild/protobuf/wire";
import { Pages } from "../../../types/v1/model";
import { RuleBased } from "./model";
export declare const protobufPackage = "sentinez.modules.security.v1";
export interface CreateRuleBasedRequest {
    ruleBased?: RuleBased | undefined;
}
export interface CreateRuleBasedResponse {
    id: string;
}
export interface GetRuleBasedRequest {
    id: string;
}
export interface GetRuleBasedResponse {
    ruleBased?: RuleBased | undefined;
}
export interface UpdateRuleBasedRequest {
    id: string;
    ruleBased?: RuleBased | undefined;
    updateMask?: string[] | undefined;
}
export interface UpdateRuleBasedResponse {
    ruleBased?: RuleBased | undefined;
}
export interface DeleteRuleBasedRequest {
    id: string;
}
export interface DeleteRuleBasedResponse {
}
export interface ListRuleBasedsRequest {
    page?: Pages | undefined;
    ids: string[];
}
export interface ListRuleBasedsResponse {
    ruleBaseds: RuleBased[];
    total: number;
}
export interface StatusRequest {
}
export interface StatusResponse {
    msg: string;
}
export declare const CreateRuleBasedRequest: MessageFns<CreateRuleBasedRequest>;
export declare const CreateRuleBasedResponse: MessageFns<CreateRuleBasedResponse>;
export declare const GetRuleBasedRequest: MessageFns<GetRuleBasedRequest>;
export declare const GetRuleBasedResponse: MessageFns<GetRuleBasedResponse>;
export declare const UpdateRuleBasedRequest: MessageFns<UpdateRuleBasedRequest>;
export declare const UpdateRuleBasedResponse: MessageFns<UpdateRuleBasedResponse>;
export declare const DeleteRuleBasedRequest: MessageFns<DeleteRuleBasedRequest>;
export declare const DeleteRuleBasedResponse: MessageFns<DeleteRuleBasedResponse>;
export declare const ListRuleBasedsRequest: MessageFns<ListRuleBasedsRequest>;
export declare const ListRuleBasedsResponse: MessageFns<ListRuleBasedsResponse>;
export declare const StatusRequest: MessageFns<StatusRequest>;
export declare const StatusResponse: MessageFns<StatusResponse>;
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
