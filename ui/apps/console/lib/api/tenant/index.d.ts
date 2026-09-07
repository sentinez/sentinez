export interface ApiOptions {
    signal?: AbortSignal;
}
export declare function getResource(id: string, options?: ApiOptions): Promise<any>;
export declare function getResourceByDomain(domain: string, options?: ApiOptions): Promise<any>;
export declare function listResources(params?: {
    resourceDomain?: string;
    resourceName?: string;
    status?: string;
    plan?: string;
}, options?: ApiOptions): Promise<any>;
//# sourceMappingURL=index.d.ts.map