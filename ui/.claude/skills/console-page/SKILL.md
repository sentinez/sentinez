---
name: console-page
description: Conventions for adding or changing pages in ui/apps/console (list / create / detail pages, sidebar tabs, loading effect, shared form components). Read BEFORE writing or refactoring any page.tsx, view.tsx or loading.tsx under ui/apps/console/app, e.g. "implement trang ...", "tạo trang list/create/edit", or a new sidebar tab.
argument-hint: '<route> (e.g. security/rate-limiter)'
---

# Console page conventions

Reference implementations: `app/console/[domain]/security/sec-rule/` and
`app/console/[domain]/security/rate-limiter/`.
For the API client the page calls, follow the `swagger-api-client` skill first.

## 1. Sidebar-tab pages MUST show the loading effect

Every page reached from the **root sidebar** (`app/console/(root)/...`, `app/tenant`)
or the **domain sidebar** (`app/console/[domain]/...`) keeps an intentional delay so
`loading.tsx` is visible when switching tabs. This is deliberate. Never remove it as
"dead code", and add it to every new sidebar-level page.

```tsx
// page.tsx (server component, no 'use client')
import View from './view';

export default async function Page() {
  await new Promise((resolve) => setTimeout(resolve, 1500));
  return <View />;
}
```

```tsx
// loading.tsx (sibling of page.tsx)
import IsLoading from '@sentinez/ui/components/common/loading';

export default function Loading() {
  return <IsLoading />;
}
```

- Use **1500 ms**, same as the existing pages (`sec-rule`, `tenant`, `member`, `analytic/*`).
- The interactive UI goes in `view.tsx` (`'use client'`); `page.tsx` only waits and renders it.
- Sub-pages that are not sidebar tabs (`new/`, `[id]/`) are client components and do not
  need the delay. `[id]/` still gets a `loading.tsx` and shows `<IsLoading />` while
  fetching.
- New sidebar entries are registered in `lib/default/index.ts`.

## 2. Route layout for a CRUD resource

```
<resource>/
  page.tsx        # 1500 ms delay -> <View />
  loading.tsx
  view.tsx        # list: @tanstack/react-table, filter, column toggle, pagination, delete dialog
  new/page.tsx    # create form -> create<X>() -> toast + router.back()
  [id]/page.tsx   # load get<X>(id) -> same form -> update<X>(id, data, UPDATE_MASK)
  [id]/loading.tsx
```

- Layout: `PageLayout`, `PageLayoutHeader` (title, subtitle, action buttons),
  `PageLayoutContent` from `@/components/page-layout`. Forms sit in
  `<div className="max-w-3xl mx-auto py-4">`.
- The list links to `./<resource>/<id>` and `./<resource>/new`.
- Table columns that read nested fields (e.g. `ingressRuntime.name`) use
  `id` + `accessorFn`, not `accessorKey`, so the name filter works.
- Status uses `statusLabel()` from `@/lib/type/security` with `BadgeStatus`.

## 3. Forms

- Put shared form fields and helpers in the section's `components/` folder and export
  them from `components/index.ts`: a `XxxFormValue` type, `createEmptyXxxForm()`,
  `xxxFormOf(apiModel)`, `xxxOf(form, id?)`, and `validateXxxForm()` that returns an
  error string or `null`.
- Reuse existing field components instead of copying them. Example: `RateLimitFields`
  wraps `SecRuleFields` and adds its own inputs.
- Validation mirrors the backend (proto `buf.validate` + service checks), and errors are
  shown with `toast.error(msg)`.
- Pages keep the form state as `useState<XxxFormValue>` with a
  `patchForm(patch) => setForm(f => ({ ...f, ...patch }))` helper.

## 4. Errors and feedback

- Wrap API calls in `try/catch` (use `catch {` without a binding when the error is
  unused), and show `toast.success` / `toast.error` from `@/lib/toast`.
- Disable the Save button while saving (`saving ? 'Saving...' : 'Save'`).

## 5. i18n (next-intl)

No user-facing string is hard-coded. Messages live in `messages/en.json` (source of
truth for key types, see `global.d.ts`) and `messages/vi.json`; add every key to both.
The locale comes from the `NEXT_LOCALE` cookie / `Accept-Language` (`i18n/locale.ts`),
there is no `/[locale]` URL segment.

- Client components: `const t = useTranslations('SecRule')`; shared words
  (Save, Cancel, Delete, Loading..., No results.) are in `Common`.
- Server components (`page.tsx` with the delay): `await getTranslations('Domain')`.
- Helpers outside React (`validateXxxForm`, `getColumns`, `handleStatusError`) take the
  translator as a parameter, typed `ReturnType<typeof useTranslations<'Validation'>>`.
- Table columns set `meta: { label: t('...') }` so the "Columns" menu is translated.
- Enum labels (status, action, priority, field source) come from
  `useSecurityOptions()` in `hooks/use-security-options.ts`, not from `lib/type/security`.
- Dates / numbers: `useFormatter()` instead of `toLocaleString()`.

## 6. Verify

```sh
cd ui/apps/console
# node from nvm if not on PATH: export PATH=$HOME/.nvm/versions/node/v26.8.1/bin:$PATH
node_modules/.bin/tsc --noEmit -p . | grep -v '\.next/'   # must be empty for your files
node_modules/.bin/eslint <changed files>
../../node_modules/.bin/prettier --write <changed files>
```

Errors that point into a stale `.next/types` cache are not caused by your change.
