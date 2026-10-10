---
name: swagger-api-client
description: Write or update a TypeScript axios API client in ui/apps/console/lib/api/<service>/ for a domain service, typed with the generated proto request/response messages from api/proto/sentinez/apps/<domain>/v1 (endpoints from the google.api.http options / api/docs/v1 swagger) and an env-gated sample-response fallback. Use when asked to "viết api", "implement API from swagger/proto", or add/change endpoints for a service (security, tenant, iam, analytic...).
argument-hint: '<service> (e.g. security)'
---

# Proto -> console API client

Reference implementation: `ui/apps/console/lib/api/security/index.ts`.
Other clients for style: `lib/api/tenant/index.ts`, `lib/api/iam/*.ts`.

**Rule: every API interface comes from `proto/sentinez/apps/<domain>/v1`.**
Import the generated types from `@sentinez/proto/sentinez/apps/<domain>/v1/<file>`
(ts-proto output in `ui/packages/proto/ts-proto`). Never hand-write interfaces that
mirror the swagger wire format, and never use the edge/runtime protos
(`dmz/edge/v1`, `types/secrule/v1`...) as API models.

## 1. Read the contract

- Service + messages: `api/proto/sentinez/apps/<domain>/v1/<domain>.proto` (rpcs,
  `google.api.http` method/path/body) and `model.proto` (models).
- `api/docs/v1/<domain>.swagger.json` is the same contract as OpenAPI; use it to
  double-check paths and query params.
- If the proto changed, regenerate first: `cd api && buf generate`.
- List the callers: `grep -rn "lib/api/<service>" apps/console`, and update them with
  the client.

## 2. File layout (`lib/api/<service>/index.ts`)

1. Imports: `axios`, the models from `.../apps/<domain>/v1/model` and every
   `XxxRequest` / `XxxResponse` from `.../apps/<domain>/v1/<domain>`.
2. `import { API_BASE_PATH } from '@/lib/api/base';` (`/api`, proxied by
   `app/api/[...path]/route.ts` to the server-only `API_BASE_PATH` env).
   `import { paginate, toQuery } from '@/lib/api/pages';` when needed.
3. `const USE_SAMPLE = process.env.NEXT_PUBLIC_USE_SAMPLE === 'true';`
4. `export interface ApiOptions { signal?: AbortSignal }`; every function takes
   `options?: ApiOptions` last and passes `signal` to axios.
5. One exported async function per rpc, with the path in a comment:

```ts
// PUT /security/secrule/{id}
export async function updateSecRule(
  req: UpdateSecRuleRequest,
  options?: ApiOptions,
): Promise<UpdateSecRuleResponse> {
  const resp = await axios.put(
    `${API_BASE_PATH}/security/secrule/${encodeURIComponent(req.id)}`,
    UpdateSecRuleRequest.toJSON(req),
    { signal: options?.signal },
  );
  return UpdateSecRuleResponse.fromJSON(resp.data ?? {});
}
```

- Signature: take `XxxRequest`, return `Promise<XxxResponse>`. Do not unwrap the
  response (callers read `.secRule`, `.id`, `.total`...). For requests with many
  optional fields (list/filter) take `Partial<XxxRequest> = {}` and normalise with
  `XxxRequest.fromPartial(req)`.
- Wire conversion is always the generated `XxxRequest.toJSON` (body) and
  `XxxResponse.fromJSON` (response). They handle enums (`STATUS_ACTIVE` <-> number),
  int64 (string <-> number), `Timestamp` (<-> `Date`), `FieldMask` (`string[]` <->
  `"a,b"`) and `bytes` (`Uint8Array` <-> base64). No manual `fromWire`/`toWire`.
- Path params: `encodeURIComponent(req.id)`.
- Query params (GET/DELETE): `params: toQuery(XxxRequest.toJSON(...))`. `toQuery`
  flattens nested messages to dotted keys (`page.index`); repeated fields stay arrays,
  so add `paramsSerializer: { indexes: null }` (`ids=a&ids=b`).
- Pagination: `Pages` comes from `@sentinez/proto/sentinez/types/v1/model`
  (re-exported by `lib/api/pages`).

## 3. Mapping to the UI

- Pages and `useApi` work with the proto models directly. Enum fields are numbers:
  compare with the enum (`Status.STATUS_ACTIVE`), display with `statusToJSON` /
  `planToJSON` or the labels in `useSecurityOptions()`.
- String fields that carry an enum name on the wire (e.g. `SecRule.action`) are
  converted with `actionTypeFromJSON` / `actionTypeToJSON` in the form helpers.
- Form <-> model mapping lives next to the form (`xxxFormOf(model)`, `xxxOf(form, id)`
  in the section's `components/`), not in `lib/api` (see the `console-page` skill).

## 4. Error handling and sample fallback

- Default: let errors throw; callers already `try/catch` + `toast.error`. (The tenant
  getters/list keep an `orNull` wrapper because `useApi` treats `null` as 404; do not
  copy that unless asked.)
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

- Pass the response type explicitly (`withFallback<GetSecRuleResponse>(...)`) so the
  fallback and the real call agree.
- Always rethrow cancellations (`axios.isCancel`).
- Exported `SAMPLE_*` data is typed with the proto models and covers each state worth
  rendering (active/disabled, AND/OR conditions...). Data easier to write as wire JSON
  can be parsed with `Model.fromJSON`. Keep a module-level `sampleStore` so
  create/update/delete are reflected in later list/get calls; build stored items with
  `Model.fromPartial` so required fields get defaults.
- Never use `Math.random`/`Date.now` at module load for anything SSR-rendered; ids
  inside sample data should be stable literals where possible.

## 5. Verify

- `cd ui/apps/console && node_modules/.bin/tsc --noEmit -p . | grep -v '\.next/'` must
  be empty, and callers must still compile.
- Do not create `.env*` files; tell the user to set `NEXT_PUBLIC_USE_SAMPLE=true` in
  `.env.local` and restart the dev server.
- Say clearly that it was not tested against the real backend unless it was.
