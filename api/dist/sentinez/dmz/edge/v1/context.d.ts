import { BinaryReader, BinaryWriter } from "@bufbuild/protobuf/wire";
import { Request } from "../../../network/http/v1/http";
import { Transport } from "../../../network/v1/conn";
import { Metadata } from "./setting";
export declare const protobufPackage = "sentinez.dmz.edge.v1";
/** Context helps the edge identify which user is connected */
export interface Context {
    metadata?: Metadata | undefined;
    transport?: Transport | undefined;
    request?: Request | undefined;
    x?: ContextExtra | undefined;
}
export interface ContextExtra {
    namespace: string;
    ruleBasedMatchedIds: string[];
    ruleBasedMatchedNames: string[];
    rulesetsMatchedIds: string[];
    rulesetsMatchedNames: string[];
    rulesetsMatchedSeverity: string[];
}
export declare const Context: MessageFns<Context>;
export declare const ContextExtra: MessageFns<ContextExtra>;
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
