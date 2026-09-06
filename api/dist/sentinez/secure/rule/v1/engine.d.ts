import { BinaryReader, BinaryWriter } from "@bufbuild/protobuf/wire";
import { Status } from "../../../types/v1/known";
export declare const protobufPackage = "sentinez.secure.rule.v1";
/** The possible sources from which a field can be extracted in a request */
export declare enum FieldSource {
    FIELD_SOURCE_UNSPECIFIED = 0,
    FIELD_SOURCE_HEADER = 1,
    FIELD_SOURCE_QUERY = 2,
    FIELD_SOURCE_PATH = 3,
    FIELD_SOURCE_BODY = 4,
    FIELD_SOURCE_IP = 5,
    FIELD_SOURCE_JA4 = 6,
    FIELD_SOURCE_TLS = 7,
    FIELD_SOURCE_METHOD = 8,
    FIELD_SOURCE_HOST = 9,
    UNRECOGNIZED = -1
}
export declare function fieldSourceFromJSON(object: any): FieldSource;
export declare function fieldSourceToJSON(object: FieldSource): string;
export declare enum Operator {
    OPERATOR_UNSPECIFIED = 0,
    OPERATOR_EQ = 1,
    OPERATOR_NE = 2,
    OPERATOR_CONTAINS = 3,
    OPERATOR_MATCHES = 4,
    OPERATOR_IN = 5,
    OPERATOR_PREFIX = 6,
    OPERATOR_SUFFIX = 7,
    OPERATOR_GT = 8,
    OPERATOR_GTE = 9,
    OPERATOR_LT = 10,
    OPERATOR_LTE = 11,
    OPERATOR_NOT_IN = 12,
    UNRECOGNIZED = -1
}
export declare function operatorFromJSON(object: any): Operator;
export declare function operatorToJSON(object: Operator): string;
export declare enum ActionType {
    ACTION_TYPE_UNSPECIFIED = 0,
    ACTION_TYPE_BLOCK = 1,
    ACTION_TYPE_LOG = 2,
    ACTION_TYPE_MODIFY_HEADER = 3,
    ACTION_TYPE_REDIRECT = 4,
    ACTION_TYPE_SET_TAG = 5,
    ACTION_TYPE_ROUTE_TO = 6,
    UNRECOGNIZED = -1
}
export declare function actionTypeFromJSON(object: any): ActionType;
export declare function actionTypeToJSON(object: ActionType): string;
/** A logical condition expression */
export interface Condition {
    /** @gotags: yaml:"-" */
    id: string;
    /**
     * The source of the field
     * Example: "User-Agent" or "country"
     */
    source: FieldSource;
    /** @gotags: yaml:"key" */
    key: string;
    /**
     * Supported operators: "eq", "ne", "contains",
     * "matches", "in", "prefix", "suffix", "gt", "lt"
     */
    operator: Operator;
    /** The value to compare against */
    value?: any | undefined;
}
/** An action to execute when a rule matches */
export interface Action {
    /**
     * Example types: "block", "log", "modify_header",
     * "redirect", "set_tag", "route_to"
     */
    type: ActionType;
    /** Dynamic parameters, e.g., { "status": 403, "message": "Forbidden" } */
    params?: {
        [key: string]: any;
    } | undefined;
}
/** A complete rule definition */
export interface Rule {
    id: string;
    /** @gotags: yaml:"name" */
    name: string;
    description: string;
    /** @gotags: yaml:"condition" */
    condition?: Condition | undefined;
}
export interface AndCondition {
    rules: Rule[];
    orCondition: AndCondition[];
}
export interface AndConditionLite {
    rules: RuleLite[];
    orCondition: AndConditionLite[];
}
/** A complete rule definition */
export interface RuleLite {
    id: string;
    /** @gotags: yaml:"name" */
    name: string;
    /** @gotags: yaml:"condition" */
    condition?: ConditionLite | undefined;
    /** @gotags: yaml:"actions" */
    actions: string[];
}
/** A logical condition expression */
export interface ConditionLite {
    /**
     * The source of the field
     * Example: "User-Agent" or "country"
     */
    source: string;
    /** @gotags: yaml:"key" */
    key: string;
    /**
     * Supported operators: "eq", "ne", "contains",
     * "matches", "in", "prefix", "suffix", "gt", "lt"
     */
    operator: string;
    /** The value to compare against */
    value: string;
}
export interface Expression {
    orCondition: AndCondition[];
}
export interface ExpressionLite {
    orCondition: AndConditionLite[];
}
export interface MatchedRules {
    ids: string[];
    names: string[];
}
export interface RuleBased {
    id: string;
    name: string;
    description: string;
    expr?: Expression | undefined;
    action?: Action | undefined;
    status: Status;
    priority: number;
    createdAt?: Date | undefined;
    updatedAt?: Date | undefined;
}
export interface RuleBasedLite {
    id: string;
    name: string;
    description: string;
    expr?: ExpressionLite | undefined;
    action?: Action | undefined;
}
export declare const Condition: MessageFns<Condition>;
export declare const Action: MessageFns<Action>;
export declare const Rule: MessageFns<Rule>;
export declare const AndCondition: MessageFns<AndCondition>;
export declare const AndConditionLite: MessageFns<AndConditionLite>;
export declare const RuleLite: MessageFns<RuleLite>;
export declare const ConditionLite: MessageFns<ConditionLite>;
export declare const Expression: MessageFns<Expression>;
export declare const ExpressionLite: MessageFns<ExpressionLite>;
export declare const MatchedRules: MessageFns<MatchedRules>;
export declare const RuleBased: MessageFns<RuleBased>;
export declare const RuleBasedLite: MessageFns<RuleBasedLite>;
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
