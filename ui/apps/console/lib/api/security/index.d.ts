export interface ApiOptions {
    signal?: AbortSignal;
}
export interface RuleBased {
    metadata?: {
        createdAt?: string;
        updatedAt?: string;
    };
    id?: string;
    name?: string;
    description?: string;
    node?: ApiRuleGroup;
    action?: any;
    status?: string;
    priority?: number;
}
export type Logic = 'LOGIC_UNSPECIFIED' | 'LOGIC_AND' | 'LOGIC_OR' | 'LOGIC_NOT';
export type FieldSource = 'FIELD_SOURCE_UNSPECIFIED' | 'FIELD_SOURCE_HEADER' | 'FIELD_SOURCE_QUERY' | 'FIELD_SOURCE_PATH' | 'FIELD_SOURCE_BODY' | 'FIELD_SOURCE_IP' | 'FIELD_SOURCE_JA4' | 'FIELD_SOURCE_TLS' | 'FIELD_SOURCE_METHOD' | 'FIELD_SOURCE_HOST';
export type Operator = 'OPERATOR_UNSPECIFIED' | 'OPERATOR_EQ' | 'OPERATOR_NE' | 'OPERATOR_CONTAINS' | 'OPERATOR_MATCHES' | 'OPERATOR_IN' | 'OPERATOR_PREFIX' | 'OPERATOR_SUFFIX' | 'OPERATOR_GT' | 'OPERATOR_GTE' | 'OPERATOR_LT' | 'OPERATOR_LTE' | 'OPERATOR_NOT_IN';
export interface ApiRule {
    id?: string;
    field?: FieldSource;
    key?: string;
    operator?: Operator;
    value?: string;
}
export interface ApiRuleGroup {
    id?: string;
    combinator?: Logic;
    rules?: ApiRuleNode[];
    not?: boolean;
}
export interface ApiRuleNode {
    rule?: ApiRule;
    group?: ApiRuleGroup;
}
export declare function transformApiToUi(apiGroup: ApiRuleGroup): any;
export declare function transformUiToApi(uiGroup: any): ApiRuleGroup;
export declare function transformApiToUiRuleBased(apiData: any): RuleBased;
export declare function transformUiToApiRuleBased(uiData: any): RuleBased;
export declare const MOCK_RULE: RuleBased;
export declare function listRuleBaseds(options?: ApiOptions): Promise<any>;
export declare function getRuleBased(id: string, options?: ApiOptions): Promise<RuleBased>;
export declare function createRuleBased(data: RuleBased, options?: ApiOptions): Promise<any>;
export declare function updateRuleBased(id: string, data: RuleBased, options?: ApiOptions): Promise<any>;
//# sourceMappingURL=index.d.ts.map