---
name: swagger-api-client
description: Write or update a TypeScript axios API client in ui/apps/console/lib/api/<service>/ from a swagger file in api/docs/v1/, mapping wire types to the proto types the UI uses and adding an env-gated sample-response fallback. Use when asked to "viết api", "implement API from swagger", or add/change endpoints for a service (security, tenant, iam, analytic...).
argument-hint: '<service> (e.g. security)'
---

# Swagger -> console API client

Reference implementation: `ui/apps/console/lib/api/security/index.ts`.
Other clients for style: `lib/api/tenant/index.ts`, `lib/api/iam/*.ts`.

## 1. Read the contract

- Spec: `api/docs/v1/<service>.swagger.json`. Go through every `paths` entry: method, path, path/query params, body schema, 200 response schema.
- Resolve `$ref`s under `definitions`. Note the wire shapes (camelCase JSON, enums as strings such as `STATUS_ACTIVE`, int64 as string, `google.rpc.Status` as the default error).
- List how callers use the existing file: `grep -rn "lib/api/<service>" apps/console`. Keep their call signatures working.

## 2. File layout (`lib/api/<service>/index.ts`)

1. Imports: `axios`, proto types from `@sentinez/proto/...` (generated in `packages/proto/ts-proto`).
2. `import { API_BASE_PATH } from '@/lib/api/base';` (`/api`, proxied by `app/api/[...path]/route.ts` to the server-only `API_BASE_PATH` env).
3. `const USE_SAMPLE = process.env.NEXT_PUBLIC_USE_SAMPLE === 'true';`
4. `export interface ApiOptions { signal?: AbortSignal }`; every function takes `options?: ApiOptions` last and passes `signal` to axios.
5. Wire types mirroring the swagger definitions (name them after the definition, e.g. `SecuritySecRule`).
6. `fromWire` / `toWire` mappers between wire types and the proto types the UI uses. Use the generated `xxxFromJSON` / `xxxToJSON` and `Message.fromJSON` / `toJSON` helpers instead of hand-rolling enum or oneof conversion. Skip mapping only if the UI already consumes the wire shape.
7. One exported async function per operation, with the path in a comment: `// GET /security/secrule/{id}`.
   - `encodeURIComponent` path params.
   - Query params: nested names use dotted keys (`'page.index'`); arrays with `collectionFormat: multi` need `paramsSerializer: { indexes: null }`.
   - Unwrap the response field (`resp.data?.secRule`), return `id` for create, `void` for delete.
   - Body wrappers follow the swagger (`{ secRule, updateMask }`).

## 3. Error handling and sample fallback

- Default: let errors throw; callers already `try/catch` + `toast.error`. (Older clients like tenant swallow errors and return null; do not copy that unless asked.)
- Sample fallback, only when `USE_SAMPLE` is true:

```ts
function withFallback<T>(label: string, fallback: () => T) {
  return async (call: () => Promise<T>): Promise<T> => {
    try {
      return await call();
    } catch (err: any) {
      if (!USE_SAMPLE || axios.isCancel(err)) throw err;
      console.warn(`[<service>] ${label} failed, using sample response:`, err?.message);
      return fallback();
    }
  };
}
```

- Always rethrow cancellations (`axios.isCancel`).
- Provide exported `SAMPLE_*` data that covers each state worth rendering (active/disabled, AND/OR conditions...). Keep a module-level `sampleStore` so create/update/delete are reflected in later list/get calls while falling back.
- Never use `Math.random`/`Date.now` at module load for anything SSR-rendered; ids inside sample data should be stable literals where possible.

## 4. Verify

- `cd ui/apps/console && npx tsc --noEmit -p . | grep "lib/api/<service>"` must be empty, and callers must still compile.
- Do not create `.env*` files; tell the user to set `NEXT_PUBLIC_USE_SAMPLE=true` in `.env.local` and restart the dev server.
- Say clearly that it was not tested against the real backend unless it was.
- If the swagger and proto types disagree (e.g. `action` is a string on the wire but an object in proto), state what is lost in mapping.
