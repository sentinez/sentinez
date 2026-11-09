# e2e

Browser automation against a running web console, built on
[Playwright](https://playwright.dev) (`playwright-core`, no bundled
browsers).

For now it holds one script, `screenshots.mjs`, which captures every console
page for the docs. It writes them to
[`docs/images/console/`](../docs/images/console/).

## Requirements

- Node 26.8.1 (the version pinned in [`ui/package.json`](../ui/package.json)).
- A Chrome or Chromium binary. The default is `/usr/bin/google-chrome`; set
  `SZ_CHROME` to use another one, e.g. a Playwright-cached Chromium.
- The console running (`cd ui && pnpm dev`, port 3000) with its backend
  (`make sz.apiserver.run`).
- An account with at least one tenant. The tenant should own a domain that
  has a SecRule, a rate limiter and a ruleset, so that the list and detail
  pages are not empty.

## Setup

```sh
cd e2e
npm install
```

## Capture the docs screenshots

```sh
SZ_USER=admin SZ_PASSWORD=<password> npm run screenshots
```

The script:

1. Captures the pages that need no session: login, sign-up and passkey
   login.
2. Logs in, captures the tenant picker, then enters the first tenant.
3. Captures the tenant-wide pages: domains, members and settings.
4. Picks a domain and captures its tenant, security, delivery and analytic
   pages. For the SecRule, rate limiter and ruleset detail pages it opens the
   first row of the list.

Each page is one PNG, 1440px wide and as tall as the page. The viewport is
grown to the page height so the sidebar is not cut off. The Next.js dev
indicator is hidden, and the script waits for the console's skeletons
(`.animate-pulse`) to go away. File names start with a number so they sort
in the sidebar's order (`10-` domains, `30-` security, `50-` analytic, ...).
A run overwrites the existing files.

| Variable | Default | Meaning |
|---|---|---|
| `SZ_USER` | (required) | Login email or username |
| `SZ_PASSWORD` | (required) | Login password |
| `SZ_BASE_URL` | `http://localhost:3000` | Console URL |
| `SZ_DOMAIN` | first domain listed | Domain used for the `/d/console/<domain>` pages |
| `SZ_OUT` | `../docs/images/console` | Output directory |
| `SZ_CHROME` | `/usr/bin/google-chrome` | Browser executable |

## Adding a page

Add a `[name, path]` entry to `PUBLIC_PAGES` (no login), `ROOT_PAGES`
(tenant-wide) or `DOMAIN_PAGES` (path relative to `/d/console/<domain>`) in
`screenshots.mjs`. For a detail page, add `'detail'` as a third element and
give the list page's path.

## Known gaps

- `analytic/activity` is still a placeholder page, so its screenshot shows
  only skeleton rows. `delivery/cdn` has a title and no content yet.
- The member detail page is not captured: the member list has no link to it.
