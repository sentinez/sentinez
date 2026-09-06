import { BinaryReader, BinaryWriter } from "@bufbuild/protobuf/wire";
export declare const protobufPackage = "sentinez.secure.rule.v1";
export declare enum RuleService {
    RULE_SERVICE_UNSPECIFIED = 0,
    RULE_SERVICE_CORE_RULESETS = 1,
    RULE_SERVICE_CUSTOM_RULES = 2,
    RULE_SERVICE_RATE_LIMIT = 3,
    UNRECOGNIZED = -1
}
export declare function ruleServiceFromJSON(object: any): RuleService;
export declare function ruleServiceToJSON(object: RuleService): string;
export declare enum RuleBehavior {
    RULE_BEHAVIOR_UNSPECIFIED = 0,
    RULE_BEHAVIOR_DENY = 1,
    UNRECOGNIZED = -1
}
export declare function ruleBehaviorFromJSON(object: any): RuleBehavior;
export declare function ruleBehaviorToJSON(object: RuleBehavior): string;
export interface CoreRulesets {
    name: string;
    rules: CoreRule[];
    version: string;
}
export interface CoreRule {
    actions?: RuleAction | undefined;
    configuration: string;
    level: string;
}
export interface RuleAction {
    statement: string;
    children?: RuleAction | undefined;
    fields?: RuleActionField | undefined;
}
export interface RuleActionField {
    id: string[];
    logdata: string[];
    msg: string[];
    phase: string[];
    setvar: string[];
    severity: string[];
    t: string[];
    ver: string[];
    tag: string[];
}
export declare const CoreRulesets: MessageFns<CoreRulesets>;
export declare const CoreRule: MessageFns<CoreRule>;
export declare const RuleAction: MessageFns<RuleAction>;
export declare const RuleActionField: MessageFns<RuleActionField>;
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
