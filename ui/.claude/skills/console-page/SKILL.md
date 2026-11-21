---
name: console-page
description: Conventions for adding or changing pages in ui/apps/console (list / create / detail pages, sidebar tabs, loading effect, centred page layout, shared form and toolbar components). Read BEFORE writing or refactoring any page.tsx, view.tsx or loading.tsx under ui/apps/console/app, e.g. "implement trang ...", "tạo trang list/create/edit", or a new sidebar tab.
argument-hint: '<route> (e.g. security/sec-rate-limit)'
---

# Console page conventions

Reference implementations: `app/d/console/[domain]/security/sec-rule/` and
`app/d/console/[domain]/security/sec-rate-limit/`.
For the API client the page calls, follow the `swagger-api-client` skill first.

## 1. Sidebar-tab pages MUST show the loading effect

Switching tabs in the **root sidebar** (`app/d/console/(root)/...`, `app/d/setting`)
or the **domain sidebar** (`app/d/console/[domain]/...`) covers the page with the
spinner for `TAB_LOADING_MILLIS` (`lib/delays.ts`). `ConsoleSidebarInset` does it through
`TabLoading` (`components/tab-loading.tsx`), so a page adds nothing for it. The page renders
and fetches under the cover. When the cover lifts, the page shows its data, or the table
skeleton if the fetch is still running. This is deliberate: never remove it. Never add a
server-side wait (`await sleep`) to `page.tsx`: it would delay the fetch instead of running
alongside it.

```tsx
// page.tsx (server component, no 'use client')
import View from './view';

export default function Page() {
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

- `loading.tsx` stays. It shows the same spinner while the server is actually slow.
- The interactive UI goes in `view.tsx` (`'use client'`); `page.tsx` only renders it.
  A page with server work awaits only that work (`const t = await getTranslations('X')`).
- Moving into a tab's own sub pages (`new/`, `[id]/`) or back out of them is not a tab
  switch, so it gets no cover. `[id]/` still gets a `loading.tsx` and shows `<IsLoading />`
  while fetching.
- New sidebar entries are registered in `lib/default/index.ts`.

## 2. Route layout for a CRUD resource

```
<resource>/
  page.tsx        # <View />
  loading.tsx
  view.tsx        # list: @tanstack/react-table, filter, column toggle, pagination, delete dialog
  new/page.tsx    # create form -> create<X>({ x: xOf(form) }) -> toast + router.back()
  [id]/page.tsx   # get<X>({ id }) -> xFormOf(res.x) -> same form
                  #   -> update<X>({ id, x: xOf(form, id), updateMask: UPDATE_MASK })
  [id]/loading.tsx
```

- Layout: `PageLayout`, `PageLayoutHeader` (title, subtitle, action buttons),
  `PageLayoutContent` from `@/components/page-layout`. The header is left-aligned and
  `max-w-7xl`; the content is centred under the header, the full width when the page is
  narrower, and its `size` sets the width from large to small: `xl` (`max-w-7xl`, the whole
  header width, for the list pages and the event log pages: histogram + event table), `lg` (`max-w-5xl`, the
  default), `md` (`max-w-3xl`) and `sm` (`max-w-xl`, for the create and update form pages `new/page.tsx`, `[id]/page.tsx`). Put the content straight in `PageLayoutContent`, never in a wrapper with its own `max-w-*`, `mx-auto` or
  `py-*` (`PageLayoutContent` already pads and spaces its children), and never pull a
  block out with a negative margin (`-mx-1`).
- Forms: `<PageLayoutContent><XxxFields ... /></PageLayoutContent>`. Short fields
  (selects, numbers) go side by side in `<FieldGrid>` (`cols={3}` for three) from
  `@/components/field-grid`, not a hand-written `grid sm:grid-cols-2`, so they do not
  stretch across the page.
- Raw JSON (a preview, a "show JSON" fold) or text (a certificate) is
  `<CodeBlock value={...} />` from `@/components/code-block`, not a hand-written
  `<pre>{JSON.stringify(...)}</pre>`: it
  keeps the parent's width (`min-w-0`) and scrolls a long line inside itself, so the card
  around it does not widen or shift when zooming.
- List: the filter / refresh / columns row above the table is `<PageLayoutToolbar>`
  (spacing included: no `ml-2` on its items, no `py-4`).
- `PageLayoutHeader` is sticky and already holds the sidebar trigger and the
  breadcrumb of the path (`BreadcrumbBar`): a page never renders its own.
- The page itself is laid out by `PageLayoutHeader` / `PageLayoutContent` only; a page
  never adds its own `mx-auto` or `items-center justify-center`. Inside a page, only
  small things centre in their own box (a table cell, a tooltip, an empty state).
- The list links to `./<resource>/<id>` and `./<resource>/new`.
- Rows are the proto models from `@sentinez/proto/sentinez/apps/<domain>/v1/model`
  (e.g. `SecRule`, `RateLimit`, `Resource`), as returned by `lib/api`. Columns that
  read nested fields use `id` + `accessorFn`, not `accessorKey`, so the name filter works.
- The table is `<FixedTable table={table}>` from `@/components/fixed-table`, not a bare
  `<Table className="table-fixed">`: columns set `size` (px), the one taking the rest
  `size: 0`, and the table scrolls instead of squeezing that column on narrow screens.
- The list fetches through `usePagedList`:
  `const res = await listX({ page }, { signal }); return { items: res.xs, total: res.total };`
  and deletes with `deleteX({ id })`.
- Enum fields are numbers: compare against the enum (`Status.STATUS_ACTIVE`); enum
  names sent as strings (`SecRule.action`) are parsed with `actionTypeFromJSON`.
- Status uses `statusLabel()` from `@/lib/type/security` with `BadgeStatus`.

## 3. Forms

- Put shared form fields and helpers in the section's `components/` folder and export
  them from `components/index.ts`: a `XxxFormValue` type, `createEmptyXxxForm()`,
  `xxxFormOf(apiModel)`, `xxxOf(form, id?)`, and `validateXxxForm()` that returns an
  error string or `null`. `apiModel` / the return of `xxxOf` is the proto model type
  (see `secRuleFormOf` / `secRuleOf` in `sec-rule-fields.tsx`).
- `UPDATE_MASK` is a `string[]` of proto field names (`['name', 'time_window']`), the
  type of `updateMask` in the generated request.
- Reuse existing field components and helpers instead of copying them. Example:
  `RateLimitFields` wraps `SecRuleFields`, and `rateLimitFormOf` / `rateLimitOf` spread
  `secRuleFormOf` / `secRuleOf`.
- Selects over proto enums (numbers) use `value={String(v)}` and
  `onValueChange={(v) => onChange(Number(v) as Plan)}`.
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
- Server components (`page.tsx`): `await getTranslations('Domain')`.
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
