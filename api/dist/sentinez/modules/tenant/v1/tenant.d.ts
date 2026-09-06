import { BinaryReader, BinaryWriter } from "@bufbuild/protobuf/wire";
import { Setting } from "../../../dmz/edge/v1/setting";
import { Plan, Status } from "../../../types/v1/known";
import { Pages } from "../../../types/v1/model";
import { Resource } from "./model";
export declare const protobufPackage = "sentinez.modules.tenant.v1";
export interface GetResourceRequest {
    id: string;
    default: boolean;
}
export interface GetResourceByDomainRequest {
    resourceDomain: string;
}
export interface GetResourceByDomainResponse {
    resource?: Resource | undefined;
}
export interface GetResourceResponse {
    resource?: Resource | undefined;
}
export interface DeleteResourceRequest {
    id: string;
}
export interface DeleteResourceResponse {
}
export interface UpdateResourceRequest {
    id: string;
    resourceDomain: string;
    resourceName: string;
    status: Status;
    plan: Plan;
}
export interface UpdateResourceResponse {
}
export interface CreateResourceRequest {
    resourceSetting?: Setting | undefined;
    resourceDomain: string;
    resourceName: string;
    status: Status;
    plan: Plan;
}
export interface CreateResourceResponse {
    resource?: Resource | undefined;
}
export interface ListResourceRequest {
    page?: Pages | undefined;
    status: Status;
    plan: Plan;
    resourceName: string;
    resourceDomain: string;
}
export interface ListResourceResponse {
    total: number;
    resources: Resource[];
}
export interface StatusResponse {
    message: string;
}
export interface StatusRequest {
}
export declare const GetResourceRequest: MessageFns<GetResourceRequest>;
export declare const GetResourceByDomainRequest: MessageFns<GetResourceByDomainRequest>;
export declare const GetResourceByDomainResponse: MessageFns<GetResourceByDomainResponse>;
export declare const GetResourceResponse: MessageFns<GetResourceResponse>;
export declare const DeleteResourceRequest: MessageFns<DeleteResourceRequest>;
export declare const DeleteResourceResponse: MessageFns<DeleteResourceResponse>;
export declare const UpdateResourceRequest: MessageFns<UpdateResourceRequest>;
export declare const UpdateResourceResponse: MessageFns<UpdateResourceResponse>;
export declare const CreateResourceRequest: MessageFns<CreateResourceRequest>;
export declare const CreateResourceResponse: MessageFns<CreateResourceResponse>;
export declare const ListResourceRequest: MessageFns<ListResourceRequest>;
export declare const ListResourceResponse: MessageFns<ListResourceResponse>;
export declare const StatusResponse: MessageFns<StatusResponse>;
export declare const StatusRequest: MessageFns<StatusRequest>;
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
