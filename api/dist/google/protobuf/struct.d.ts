import { BinaryReader, BinaryWriter } from "@bufbuild/protobuf/wire";
export declare const protobufPackage = "google.protobuf";
/**
 * `NullValue` is a singleton enumeration to represent the null value for the
 * `Value` type union.
 *
 * The JSON representation for `NullValue` is JSON `null`.
 */
export declare enum NullValue {
    /** NULL_VALUE - Null value. */
    NULL_VALUE = 0,
    UNRECOGNIZED = -1
}
export declare function nullValueFromJSON(object: any): NullValue;
export declare function nullValueToJSON(object: NullValue): string;
/**
 * `Struct` represents a structured data value, consisting of fields
 * which map to dynamically typed values. In some languages, `Struct`
 * might be supported by a native representation. For example, in
 * scripting languages like JS a struct is represented as an
 * object. The details of that representation are described together
 * with the proto support for the language.
 *
 * The JSON representation for `Struct` is JSON object.
 */
export interface Struct {
    /** Unordered map of dynamically typed values. */
    fields: {
        [key: string]: any | undefined;
    };
}
export interface Struct_FieldsEntry {
    key: string;
    value?: any | undefined;
}
/**
 * `Value` represents a dynamically typed value which can be either
 * null, a number, a string, a boolean, a recursive struct value, or a
 * list of values. A producer of value is expected to set one of these
 * variants. Absence of any variant indicates an error.
 *
 * The JSON representation for `Value` is JSON value.
 */
export interface Value {
    /** Represents a null value. */
    nullValue?: NullValue | undefined;
    /** Represents a double value. */
    numberValue?: number | undefined;
    /** Represents a string value. */
    stringValue?: string | undefined;
    /** Represents a boolean value. */
    boolValue?: boolean | undefined;
    /** Represents a structured value. */
    structValue?: {
        [key: string]: any;
    } | undefined;
    /** Represents a repeated `Value`. */
    listValue?: Array<any> | undefined;
}
/**
 * `ListValue` is a wrapper around a repeated field of values.
 *
 * The JSON representation for `ListValue` is JSON array.
 */
export interface ListValue {
    /** Repeated field of dynamically typed values. */
    values: any[];
}
export declare const Struct: MessageFns<Struct> & StructWrapperFns;
export declare const Struct_FieldsEntry: MessageFns<Struct_FieldsEntry>;
export declare const Value: MessageFns<Value> & AnyValueWrapperFns;
export declare const ListValue: MessageFns<ListValue> & ListValueWrapperFns;
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
export interface StructWrapperFns {
    wrap(object: {
        [key: string]: any;
    } | undefined): Struct;
    unwrap(message: Struct): {
        [key: string]: any;
    };
}
export interface AnyValueWrapperFns {
    wrap(value: any): Value;
    unwrap(message: any): string | number | boolean | Object | null | Array<any> | undefined;
}
export interface ListValueWrapperFns {
    wrap(array: Array<any> | undefined): ListValue;
    unwrap(message: ListValue): Array<any>;
}
export {};
