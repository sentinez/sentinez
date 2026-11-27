# recept

A package index in the spirit of pkg.go.dev for the security rules of the
Sentinéz edge, built with Astro and the shadcn components of `@sentinez/ui`.
Every page is static: `pnpm build` writes `dist/`.

```sh
pnpm dev      # http://localhost:4321
pnpm build && pnpm preview
```

## Adding a package

A package is a folder of `pkg/` holding YAML and Lua files only; its path is
the folder path (`pkg/auth/login` → `auth/login`). The page is read from the
files, as pkg.go.dev reads Go: no README.

- **YAML**: the `security` block of `proxy.yaml` (`cmd/szedge/v1`), with or
  without its `security:` key: `secrules`, `limiters`, `ruleset`. Each rule is
  documented by the comment right above it; the page shows its condition as
  text (`path == "/login" and method == "GET"`), its action, its limit, or the
  settings of the ruleset.
- **Lua**: a rule script (`match(ctx)` for a secrule, `key(ctx)` for a rate
  limiter) or hooks (`---@stage http.request/<stage>`), see
  `staging/src/github.com/sentinez/scripting/README.md`. Each global function
  is documented by its `---` comment lines.

The package doc is the first comment block of a file, when a blank line
follows it (else it documents the rule below). Its first sentence is the
synopsis; markdown is rendered; `@tag value` lines set the metadata:

```yaml
# Protects a login page: a challenge before the form, and a rate limit on
# the sign-in requests.
#
# Longer description, in markdown.
#
# @version v2.0.0
# @license Apache-2.0
# @keywords login, brute-force

secrules:
  # The form: GET only.
  - ingress:
      name: Challenge GET /login
      ...
```

In Lua the same block is written with `--`. The "Updated" date is the last
commit of the folder (its newest file when uncommitted). The parsers live in
`src/lib/rule.ts`, the loader in `src/lib/pkg.ts`.
