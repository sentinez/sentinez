import { BinaryReader, BinaryWriter } from "@bufbuild/protobuf/wire";
import { Setting } from "../../../dmz/edge/v1/setting";
import { Plan, Status } from "../../../types/v1/known";
import { Metadata } from "../../../types/v1/model";
export declare const protobufPackage = "sentinez.modules.tenant.v1";
export interface Resource {
    metadata?: Metadata | undefined;
    id: string;
    resourceSetting?: Setting | undefined;
    resourceDomain: string;
    resourceName: string;
    status: Status;
    plan: Plan;
}
export declare const Resource: MessageFns<Resource>;
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
