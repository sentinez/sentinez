import { BinaryReader, BinaryWriter } from "@bufbuild/protobuf/wire";
export declare const protobufPackage = "sentinez.types.v1";
export declare enum Kind {
    KIND_UNSPECIFIED = 0,
    KIND_ACCESS_ZONE = 1,
    KIND_DEMILITARIZED_ZONE = 2,
    KIND_MESH = 4,
    UNRECOGNIZED = -1
}
export declare function kindFromJSON(object: any): Kind;
export declare function kindToJSON(object: Kind): string;
export declare enum Console {
    CONSOLE_UNSPECIFIED = 0,
    CONSOLE_PORTAL = 1,
    CONSOLE_ADMIN = 2,
    UNRECOGNIZED = -1
}
export declare function consoleFromJSON(object: any): Console;
export declare function consoleToJSON(object: Console): string;
export declare enum Status {
    STATUS_UNSPECIFIED = 0,
    STATUS_ACTIVE = 1,
    STATUS_DISABLE = 2,
    UNRECOGNIZED = -1
}
export declare function statusFromJSON(object: any): Status;
export declare function statusToJSON(object: Status): string;
export declare enum Plan {
    PLAN_UNSPECIFIED = 0,
    PLAN_FREE = 1,
    PLAN_STANDARD = 2,
    PLAN_PRO = 3,
    UNRECOGNIZED = -1
}
export declare function planFromJSON(object: any): Plan;
export declare function planToJSON(object: Plan): string;
export declare enum LogKind {
    LOG_KIND_UNSPECIFIED = 0,
    LOG_KIND_HTTP = 1,
    LOG_KIND_WAF = 2,
    LOG_KIND_RULE = 3,
    UNRECOGNIZED = -1
}
export declare function logKindFromJSON(object: any): LogKind;
export declare function logKindToJSON(object: LogKind): string;
export declare enum Errors {
    ERRORS_UNSPECIFIED = 0,
    ERRORS_INTERNAL_ERROR = 1,
    ERRORS_NOT_FOUND = 2,
    ERRORS_UNAUTHORIZED = 3,
    ERRORS_FORBIDDEN = 4,
    ERRORS_INVALID_DATA = 5,
    ERRORS_UNIMPLEMENTED = 6,
    UNRECOGNIZED = -1
}
export declare function errorsFromJSON(object: any): Errors;
export declare function errorsToJSON(object: Errors): string;
export interface Empty {
}
export interface Context {
    name: string;
    expireAt?: Date | undefined;
    userId: string;
    console: Console;
}
export declare const Empty: MessageFns<Empty>;
export declare const Context: MessageFns<Context>;
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
