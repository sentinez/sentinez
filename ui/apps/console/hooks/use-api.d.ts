/**
 * A standard hook for managing data fetching with automatic error toasts and request cancellation.
 */
export declare function useApi<T>(apiFn: (...args: any[]) => Promise<T>, ...args: any[]): {
    data: T | null;
    isLoading: boolean;
    error: any;
    refresh: (signal?: AbortSignal) => Promise<void>;
};
//# sourceMappingURL=use-api.d.ts.map