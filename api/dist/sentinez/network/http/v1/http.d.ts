import { BinaryReader, BinaryWriter } from "@bufbuild/protobuf/wire";
export declare const protobufPackage = "sentinez.network.http.v1";
export interface RequestHeader {
    key: Uint8Array;
    values: Uint8Array[];
}
export interface RequestQuery {
    key: Uint8Array;
    values: Uint8Array[];
}
export interface Request {
    id: string;
    method: string;
    scheme: string;
    host: string;
    path: Uint8Array;
    uri: Uint8Array;
    queries: RequestQuery[];
    headers: RequestHeader[];
    body: Uint8Array;
    protocol: string;
    remoteAddress: string;
    status: number;
    userAgent: Uint8Array;
    rayId: string;
    asn: number;
    country: string;
    contentType: Uint8Array;
    timestamp?: Date | undefined;
    fingerprint: string;
    clientIp: string;
}
export interface HeaderValue {
    values: string[];
}
export interface QueryValue {
    values: string[];
}
export interface RequestEvent {
    id: string;
    method: string;
    scheme: string;
    host: string;
    path: string;
    queries: {
        [key: string]: QueryValue;
    };
    headers: {
        [key: string]: HeaderValue;
    };
    protocol: string;
    remoteAddress: string;
    status: number;
    userAgent: string;
    rayId: string;
    asn: number;
    country: string;
    contentType: string;
    timestamp: number;
    fingerprint: string;
    clientIp: string;
}
export interface RequestEvent_QueriesEntry {
    key: string;
    value?: QueryValue | undefined;
}
export interface RequestEvent_HeadersEntry {
    key: string;
    value?: HeaderValue | undefined;
}
export declare const RequestHeader: MessageFns<RequestHeader>;
export declare const RequestQuery: MessageFns<RequestQuery>;
export declare const Request: MessageFns<Request>;
export declare const HeaderValue: MessageFns<HeaderValue>;
export declare const QueryValue: MessageFns<QueryValue>;
export declare const RequestEvent: MessageFns<RequestEvent>;
export declare const RequestEvent_QueriesEntry: MessageFns<RequestEvent_QueriesEntry>;
export declare const RequestEvent_HeadersEntry: MessageFns<RequestEvent_HeadersEntry>;
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
