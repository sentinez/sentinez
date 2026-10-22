# github.com/sentinez/go-import

> ⚠️ **This is an automatically published [staged repository](https://github.com/sentinez/go-import) for Sentinez**.   
> Contributions, including issues and pull requests, should be made to the main Sentinez repository: [https://github.com/sentinez/sentinez](https://github.com/sentinez/sentinez).  
> This repository is read-only for importing, and not used for direct contributions.

=======================================================================

Vanity import server for `go.sentinez.com`, serving `go-import` meta tags
that map module paths to Sentinéz GitHub repos. It runs as a Vercel Go
serverless function.

## How it works

The first path segment names the repo; everything after it is a package
inside that repo:

```
go.sentinez.com/core/szhttp/chains  ->  https://github.com/sentinez/core
go.sentinez.com/shared/szlog        ->  https://github.com/sentinez/shared
go.sentinez.com/contrib/httphz      ->  https://github.com/sentinez/contrib
```

| Request | Response |
|---|---|
| `GET /<repo>/...?go-get=1` | `go-import` and `go-source` meta tags, used by `go get` / `go mod tidy` |
| `GET /` | Project page listing all modules |
| `GET /<repo>/...` | Repo page with the import path, `go get` command, docs and source links |
| `GET /_static/*.png` | Logo and favicon |
| unknown repo | `404` |

Only repos in the allowlist (`_defaultRepos` in
[vanity/vanity.go](vanity/vanity.go)) are served: `sentinez`, `core`,
`shared`, `controlplane`, `contrib`, `bpf`, `tools`, `mcp`, `go-import`.

Check a path with:

```sh
curl 'https://go.sentinez.com/core?go-get=1'
```

## Layout

| Path | Purpose |
|---|---|
| `api/index.go` | Vercel entrypoint (`Handler`) |
| `vanity/` | Request handling, repo allowlist and config |
| `vanity/view.templ` | HTML views ([templ](https://templ.guide)); `view_templ.go` is generated |
| `vanity/static/` | Embedded logo and favicon |
| `vercel.json` | Rewrites every path to `/api/index` |

## Development

```sh
go generate ./...   # regenerate view_templ.go after editing view.templ
go test ./...
```

`go generate` runs `templ` v0.3.1020 through `go run`, matching the
library version in `go.mod`. Commit the generated `view_templ.go`;
Vercel does not run `templ generate`.

To add a repo, append it to `_defaultRepos` in
[vanity/vanity.go](vanity/vanity.go).

## Deployment

1. Import this repo into Vercel (from the monorepo, set Root Directory to
   `staging/src/github.com/sentinez/go-import`). The Go version comes from
   `go.mod`.
2. Add the `go.sentinez.com` domain to the project and point a `CNAME`
   for `go` at `cname.vercel-dns.com`.
3. Module paths in each repo's `go.mod` must be `go.sentinez.com/<repo>/...`,
   otherwise `go get` fails with "module declares its path as ...".
   Modules in a subdirectory (e.g. `contrib/httphz`) need prefixed tags
   such as `httphz/v0.1.0`.

## License

Code is licensed under the [Apache License 2.0](LICENSE).

The Sentinéz logo files in `vanity/static/` are licensed under
[CC-BY-4.0](vanity/static/LICENSE).
