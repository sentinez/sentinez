import { BinaryReader, BinaryWriter } from "@bufbuild/protobuf/wire";
import { Expression } from "../../../secure/rule/v1/engine";
import { Status } from "../../../types/v1/known";
export declare const protobufPackage = "sentinez.cdn.rule.v1";
export interface Rule {
    id: string;
    name: string;
    description: string;
    expr?: Expression | undefined;
    status: Status;
    priority: number;
    createdAt?: Date | undefined;
    updatedAt?: Date | undefined;
}
export interface CDN {
    /** @gotags: yaml:"enable" */
    enable: boolean;
    /** @gotags: yaml:"rule" */
    rule?: Rule | undefined;
}
export declare const Rule: MessageFns<Rule>;
export declare const CDN: MessageFns<CDN>;
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
