import { BinaryReader, BinaryWriter } from "@bufbuild/protobuf/wire";
import { XMeta } from "../../types/v1/options";
import { Flag } from "./flags";
export declare const protobufPackage = "sentinez.setting.v1";
export declare enum Senz {
    SENZ_UNSPECIFIED = 0,
    /** SENZ_HOSTNAME - hostname of service. eg s6z.io.vn */
    SENZ_HOSTNAME = 1,
    /** SENZ_ADDRESS - http address, api, application ... */
    SENZ_ADDRESS = 2,
    /** SENZ_SECRET_KEY - secret key base64 encoded */
    SENZ_SECRET_KEY = 3,
    /** SENZ_CLIENT_ORIGIN - origin of client, used to gen passkey */
    SENZ_CLIENT_ORIGIN = 4,
    SENZ_TIMESCALE_URI = 5,
    SENZ_POSTGRES_URI = 6,
    SENZ_CLICKHOUSE_URI = 7,
    SENZ_CONSUL_URI = 8,
    SENZ_MEMBERSHIP_ADDRESS = 9,
    SENZ_DISCOVERY_ADDRESS = 10,
    UNRECOGNIZED = -1
}
export declare function senzFromJSON(object: any): Senz;
export declare function senzToJSON(object: Senz): string;
export interface Config {
    meta?: XMeta | undefined;
    flag?: Flag | undefined;
    env: {
        [key: string]: string;
    };
}
export interface Config_EnvEntry {
    key: string;
    value: string;
}
export declare const Config: MessageFns<Config>;
export declare const Config_EnvEntry: MessageFns<Config_EnvEntry>;
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
