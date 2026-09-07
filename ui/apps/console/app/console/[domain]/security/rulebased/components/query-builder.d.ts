import React from 'react';
export type Combinator = 'and' | 'or';
export type Operator = '==' | '!=' | '>' | '<' | '>=' | '<=' | 'contains' | 'startsWith' | 'endsWith' | 'matches' | 'in';
export interface Rule {
    id: string;
    field: string;
    operator: Operator;
    value: string;
}
export interface RuleGroup {
    id: string;
    combinator: Combinator;
    rules: (Rule | RuleGroup)[];
    not?: boolean;
}
interface QueryBuilderProps {
    initialQuery?: RuleGroup;
    onChange?: (query: RuleGroup) => void;
    layout?: 'vertical' | 'horizontal';
}
export declare function QueryBuilder({ initialQuery, onChange, layout }: QueryBuilderProps): React.JSX.Element;
export {};
//# sourceMappingURL=query-builder.d.ts.map