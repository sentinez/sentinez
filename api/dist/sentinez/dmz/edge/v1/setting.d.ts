import { BinaryReader, BinaryWriter } from "@bufbuild/protobuf/wire";
import { CDN } from "../../../cdn/rule/v1/cdn";
import { RuleBased as RuleBased1, RuleBasedLite } from "../../../secure/rule/v1/engine";
export declare const protobufPackage = "sentinez.dmz.edge.v1";
export declare enum BalanceStrategy {
    BALANCE_STRATEGY_UNSPECIFIED = 0,
    BALANCE_STRATEGY_ROUND_ROBIN = 1,
    UNRECOGNIZED = -1
}
export declare function balanceStrategyFromJSON(object: any): BalanceStrategy;
export declare function balanceStrategyToJSON(object: BalanceStrategy): string;
export declare enum ProxyProtocol {
    PROXY_PROTOCOL_UNSPECIFIED = 0,
    PROXY_PROTOCOL_HTTP = 1,
    PROXY_PROTOCOL_HTTPS = 2,
    UNRECOGNIZED = -1
}
export declare function proxyProtocolFromJSON(object: any): ProxyProtocol;
export declare function proxyProtocolToJSON(object: ProxyProtocol): string;
/** Setting edge setting per user */
export interface Setting {
    /** @gotags: yaml:"metadata" */
    metadata?: Metadata | undefined;
    /** @gotags: yaml:"server" */
    server?: Server | undefined;
    /** @gotags: yaml:"security" */
    security?: Security | undefined;
    /** @gotags: yaml:"controller" */
    controller?: Controller | undefined;
    /** @gotags: yaml:"personal" */
    personal?: Personal | undefined;
}
/** Metadata used for observability and debugging: */
export interface Metadata {
}
/** Server defines where the request goes and how the edge processes it: */
export interface Server {
    /** @gotags: yaml:"name" */
    name: string;
    /** @gotags: yaml:"locations" */
    locations: Location[];
}
export interface Location {
    /** @gotags: yaml:"location" */
    location: string;
    /** @gotags: yaml:"proxyRewrite" */
    proxyRewrite: string;
    /** @gotags: yaml:"proxyPass" */
    proxyPass: Upstream[];
    /** @gotags: yaml:"balanceStrategy" */
    balanceStrategy: BalanceStrategy;
    /** @gotags: yaml:"proxySetHeaders" */
    proxySetHeaders: {
        [key: string]: string;
    };
}
export interface Location_ProxySetHeadersEntry {
    key: string;
    value: string;
}
/** Security user-specific WAF, rate limiting, or bot protection rules */
export interface Security {
    /** @gotags: yaml:"rulesets" */
    rulesets: Rulesets[];
    /** @gotags: yaml:"rules" */
    rules: RuleBased[];
    /** @gotags: yaml:"limiters" */
    limiters: RateLimit[];
}
export interface Rulesets {
    enable: boolean;
}
export interface RuleBased {
    enable: boolean;
    /** @gotags: yaml:"ingress" */
    ingress?: RuleBasedLite | undefined;
    /** @gotags: yaml:"-" */
    ingressCompiled?: RuleBased1 | undefined;
}
export interface RateLimit {
    /** @gotags: yaml:"isRateLimitOn" */
    enable: boolean;
    /** @gotags: yaml:"timeWindow" */
    timeWindow: string;
    /** @gotags: yaml:"limit" */
    limit: number;
    /** @gotags: yaml:"timeout" */
    timeout: string;
}
/** Controller for systems using a virtual waiting room or throttling: */
export interface Controller {
    /** @gotags: yaml:"cdn" */
    cdn: CDN[];
}
/** Personal defines where the request goes and how the edge processes it */
export interface Personal {
}
export interface Upstream {
    /** @gotags: yaml:"server" */
    server: string;
    /** @gotags: yaml:"protocol" */
    protocol: ProxyProtocol;
}
export declare const Setting: MessageFns<Setting>;
export declare const Metadata: MessageFns<Metadata>;
export declare const Server: MessageFns<Server>;
export declare const Location: MessageFns<Location>;
export declare const Location_ProxySetHeadersEntry: MessageFns<Location_ProxySetHeadersEntry>;
export declare const Security: MessageFns<Security>;
export declare const Rulesets: MessageFns<Rulesets>;
export declare const RuleBased: MessageFns<RuleBased>;
export declare const RateLimit: MessageFns<RateLimit>;
export declare const Controller: MessageFns<Controller>;
export declare const Personal: MessageFns<Personal>;
export declare const Upstream: MessageFns<Upstream>;
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
