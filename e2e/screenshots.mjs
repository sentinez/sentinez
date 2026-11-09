// Captures every console page as a full-height PNG for the docs.
// Usage: SZ_USER=... SZ_PASSWORD=... node screenshots.mjs
// See README.md for the environment variables.

import { mkdir } from 'node:fs/promises';
import { chromium } from 'playwright-core';

const BASE = process.env.SZ_BASE_URL ?? 'http://localhost:3000';
const USER = process.env.SZ_USER;
const PASSWORD = process.env.SZ_PASSWORD;
const OUT = process.env.SZ_OUT ?? '../docs/images/console';
const CHROME = process.env.SZ_CHROME ?? '/usr/bin/google-chrome';
const WIDTH = 1440;

// Hides the Next.js dev indicator and build watcher.
const HIDE_DEV = 'nextjs-portal,[data-nextjs-toast],#__next-build-watcher'
  + '{display:none!important}';

const PUBLIC_PAGES = [
  ['01-auth-login', '/auth'],
  ['02-auth-signup', '/auth/signup'],
  ['03-auth-passkey-login', '/auth/passkey/login'],
];

const ROOT_PAGES = [
  ['10-domain-list', '/d/console/domain'],
  ['11-domain-new', '/d/console/domain/new'],
  ['12-member-list', '/d/console/member'],
  ['13-member-new', '/d/console/member/new'],
  ['60-setting-account', '/d/setting/account'],
  ['61-setting-tenant', '/d/setting/tenant'],
];

// Pages under /d/console/<domain>; `detail` pages open the first row.
const DOMAIN_PAGES = [
  ['20-tenant-resource', '/tenant/resource'],
  ['21-tenant-origin', '/tenant/origin'],
  ['22-tenant-setting', '/tenant/setting'],
  ['30-security-sec-rule-list', '/security/sec-rule'],
  ['31-security-sec-rule-new', '/security/sec-rule/new'],
  ['32-security-sec-rule-detail', '/security/sec-rule', 'detail'],
  ['33-security-rate-limiter-list', '/security/rate-limiter'],
  ['34-security-rate-limiter-new', '/security/rate-limiter/new'],
  ['35-security-rate-limiter-detail', '/security/rate-limiter', 'detail'],
  ['36-security-rulesets-list', '/security/rulesets'],
  ['37-security-rulesets-new', '/security/rulesets/new'],
  ['38-security-rulesets-detail', '/security/rulesets', 'detail'],
  ['40-delivery-cdn', '/delivery/cdn'],
  ['50-analytic-activity', '/analytic/activity'],
  ['52-analytic-logs-ingress', '/analytic/logs/ingress'],
  ['53-analytic-logs-egress', '/analytic/logs/egress'],
  ['54-analytic-logs-sec-rule', '/analytic/logs/sec-rule'],
  ['55-analytic-logs-rulesets', '/analytic/logs/rulesets'],
  ['56-analytic-logs-rate-limiter', '/analytic/logs/rate-limiter'],
];

async function open(ctx, path) {
  const page = await ctx.newPage();
  await page.goto(BASE + path, { waitUntil: 'networkidle' })
    .catch(() => {});
  await page.addStyleTag({ content: HIDE_DEV });
  // The console sidebar waits 1500ms before rendering.
  await page.waitForTimeout(2500);
  await page.waitForFunction(
    () => !document.querySelector('.animate-pulse'),
    null,
    { timeout: 20000 },
  ).catch(() => console.warn(`${path}: still loading, captured anyway`));
  return page;
}

async function shoot(ctx, name, path) {
  const page = await open(ctx, path);
  // The sidebar is viewport-high, so grow the viewport instead of fullPage.
  const height = await page.evaluate(
    () => document.documentElement.scrollHeight,
  );
  await page.setViewportSize({ width: WIDTH, height: Math.max(900, height) });
  await page.waitForTimeout(800);
  await page.screenshot({ path: `${OUT}/${name}.png` });
  console.log(`${name} <- ${page.url()}`);
  await page.close();
}

// Returns the link paths on the page at path.
async function links(ctx, path) {
  const page = await open(ctx, path);
  const hrefs = await page.$$eval('a', (as) => as.map((a) => a.href));
  await page.close();
  return hrefs.map((h) => new URL(h).pathname);
}

// Returns the first detail link of a list page (its rows link to path/<id>).
async function firstDetail(ctx, path) {
  return (await links(ctx, path))
    .find((p) => p.startsWith(path + '/') && !p.endsWith('/new'));
}

// Returns the first domain listed on the domain page.
async function firstDomain(ctx) {
  return (await links(ctx, '/d/console/domain'))
    .map((p) => p.match(/^\/d\/console\/([^/]+)\/tenant\//)?.[1])
    .find(Boolean);
}

async function login(browser) {
  const ctx = await browser.newContext({ viewport: { width: WIDTH, height: 900 } });
  const page = await ctx.newPage();
  await page.goto(BASE + '/auth');
  await page.fill('#email', USER);
  await page.fill('#password', PASSWORD);
  await page.click('button[type=submit]');
  await page.waitForURL('**/tenant');
  await page.addStyleTag({ content: HIDE_DEV });
  await page.waitForTimeout(2500);
  await page.screenshot({ path: `${OUT}/04-tenant-select.png` });
  console.log('04-tenant-select <- /tenant');
  // Enter the first tenant.
  await page.getByRole('button', { name: 'Go' }).first().click();
  await page.waitForURL('**/d/console/**');
  await page.close();
  return ctx;
}

async function main() {
  if (!USER || !PASSWORD) {
    throw new Error('set SZ_USER and SZ_PASSWORD');
  }
  await mkdir(OUT, { recursive: true });
  const browser = await chromium.launch({ executablePath: CHROME });
  try {
    const anon = await browser.newContext({
      viewport: { width: WIDTH, height: 900 },
    });
    for (const [name, path] of PUBLIC_PAGES) await shoot(anon, name, path);

    const ctx = await login(browser);
    for (const [name, path] of ROOT_PAGES) await shoot(ctx, name, path);

    const domain = process.env.SZ_DOMAIN ?? await firstDomain(ctx);
    if (!domain) throw new Error('no domain found; set SZ_DOMAIN');
    const root = `/d/console/${domain}`;

    for (const [name, path, kind] of DOMAIN_PAGES) {
      if (kind !== 'detail') {
        await shoot(ctx, name, root + path);
        continue;
      }
      const first = await firstDetail(ctx, root + path);
      if (!first) {
        console.warn(`${name}: no rows at ${path}, skipped`);
        continue;
      }
      await shoot(ctx, name, first);
    }
  } finally {
    await browser.close();
  }
}

main().catch((err) => {
  console.error(err.message);
  process.exit(1);
});
