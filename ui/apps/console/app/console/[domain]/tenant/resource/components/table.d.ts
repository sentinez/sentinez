import * as React from 'react';
import { type ColumnDef } from '@tanstack/react-table';
export type Payment = {
    id: string;
    plan: string;
    status: 'active' | 'disable' | 'success' | 'failed';
    name: string;
};
export declare const columns: ColumnDef<Payment>[];
export declare function ResourceTable(): React.JSX.Element;
//# sourceMappingURL=table.d.ts.map