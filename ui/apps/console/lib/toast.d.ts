/**
 * Enhanced toast utility for the console application.
 * Wraps sonner toast with a consistent API and safe client-side execution.
 */
export declare const toast: {
    success: (title: string, description?: string) => string | number;
    error: (title: string, description?: string) => string | number;
    info: (title: string, description?: string) => string | number;
    warning: (title: string, description?: string) => string | number;
    loading: (title: string, description?: string) => string | number;
    dismiss: (id?: string | number) => string | number;
};
//# sourceMappingURL=toast.d.ts.map