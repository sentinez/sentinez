# 13. Release

How a version of Sentinéz is cut. It is written for 0.0.1, the first
release; later releases follow the same steps. What changed goes into
[CHANGELOG.md](../CHANGELOG.md), and what to test goes into
[14-test-plan.md](14-test-plan.md).

## 13.1 Versions

- [SemVer](https://semver.org/). Before 1.0.0 there is no compatibility
  promise: the protos, `proxy.yaml`, the REST API and the schema may change
  in any release.
- One version for everything. The repo, the staged modules, the images and
  the console all ship at the same number.
- `core.Version`
  ([core/core.go](../staging/src/github.com/sentinez/core/core.go)) is the
  version shown in the startup banner. It is a `const`, `v0.0.1-beta` today,
  so `-ldflags -X` cannot set it. Set it by hand for the release, or turn it
  into a `var` and pass `-X github.com/sentinez/core.Version=...` from the
  Makefiles and Dockerfiles.

| What | Where | Tag / version |
|---|---|---|
| Root module `github.com/sentinez/sentinez` | this repo | `v0.0.1` |
| `github.com/sentinez/sentinez/api` | this repo, `api/` | `api/v0.0.1` |
| `core`, `pkg`, `controlplane`, `tools`, `scripting`, `bpf`, `integration` | split repos `github.com/sentinez/<name>` | `v0.0.1` |
| `contrib/pkg`, `contrib/httphz` | split repo `github.com/sentinez/contrib` | `pkg/v0.0.1`, `httphz/v0.0.1` |
| `go-import` | split repo | not tagged (it is the vanity import server, not a library) |
| Images | `sentinez/{apiserver,edge,greeter,centraldata,collector}` | `v0.0.1` |
| Console, web | `ui/package.json`, `ui/apps/{console,web}/package.json` | `0.0.1` (already set) |

## 13.2 Before tagging: the go.mod of the staged modules

Inside this repo every module reaches the others through `replace` with a
relative path. Go ignores the `replace` directives of a dependency, so a
staged module fetched by itself resolves its `require` lines as written:

| Module | Requires that do not resolve outside the repo |
|---|---|
| `core` | `contrib/pkg v0.0.0-00010101000000-…`, `pkg v0.0.0-00010101000000-…`, `sentinez/api v0.0.0` |
| `pkg` | `sentinez/api v0.0.0` |
| `controlplane` | `core`, `pkg` pseudo-versions of commits the split repos may not have, `sentinez/api v0.0.0`, `contrib/pkg v0.0.0-00010101000000-…` |
| `contrib/httphz`, `tools` | the same as `core` |

`contrib/pkg`, `scripting`, `bpf` and `integration` have no Sentinéz
dependency and work as they are.

Before tagging, set each of these requires to `v0.0.1`. The `replace`
directives stay, so the local build does not change. Then run `go mod tidy`
in each module and at the root. Tag in dependency order, so that every
version a module requires already exists when it is fetched:

1. `api/v0.0.1` (this repo), `pkg/v0.0.1` (the `contrib` repo),
   `scripting`, `bpf`, `integration`
2. `pkg`
3. `core`
4. `httphz/v0.0.1` (the `contrib` repo), `tools`, `controlplane`
5. root `v0.0.1`

## 13.3 Checklist

1. **Freeze.** Merge what goes into the release. Branch `release/0.0.1`
   from `main`; CI lint and test also run on `release/*`.
2. **Gates.** Run `make release.check` and make sure it passes. It checks
   gofmt, runs golangci-lint and `go test` in every module of
   [hack/gomodules.sh](../hack/gomodules.sh), runs `buf lint`, and builds
   every service. Then run `pnpm lint && pnpm typecheck && pnpm build` in
   `ui/`.
3. **Test plan.** Run [14-test-plan.md](14-test-plan.md) on a fresh
   environment and record the results in its table.
4. **Known issues.** Review [12-known-issues.md](12-known-issues.md). Either
   fix each security entry (S4 to S8) or list it in the release notes. Do
   not publish the edge image while it still bakes in the certificate of the
   source tree (S7).
5. **Version.** Set `core.Version` to `v0.0.1` and update the go.mod files as
   described in 13.2. In [CHANGELOG.md](../CHANGELOG.md), replace
   `Unreleased` with the date.
6. **Merge** the release branch into `main` (PR). `ci_sync_repo` pushes each
   staged module to the `main` branch of its split repo.
7. **Tag** in the order of 13.2. The sync pushes branches, not tags. Tag
   each split repo on the commit the sync pushed, for example:

   ```sh
   git clone git@github.com:sentinez/core.git && cd core
   git tag -a v0.0.1 -m "v0.0.1" origin/main && git push origin v0.0.1
   ```

   In this repo, tag the merge commit:

   ```sh
   git tag -a api/v0.0.1 -m "api v0.0.1"
   git tag -a v0.0.1 -m "v0.0.1"
   git push origin api/v0.0.1 v0.0.1
   bash hack/gitlab_sync.sh     # the GitLab mirror, tags included
   ```

8. **Images.** `sz.<svc>.image.build` always tags `$(TAG):latest` and has
   no version tag, so add the version tag yourself:

   ```sh
   make sz.edge.image.build
   docker tag sentinez/edge:latest sentinez/edge:v0.0.1
   docker push sentinez/edge:v0.0.1 && docker push sentinez/edge:latest
   ```

   Repeat for apiserver, greeter, centraldata and collector.
9. **GitHub release.** Create the release on `v0.0.1`. Its notes are the
   0.0.1 section of the changelog.

## 13.4 After the release

- Check each module through the proxy:
  `GOPROXY=https://proxy.golang.org go list -m github.com/sentinez/<name>@v0.0.1`.
- In an empty module, run `go get github.com/sentinez/core@v0.0.1` and then
  `go build ./...`. This catches any require that 13.2 missed.
- Pull the `v0.0.1` images. Start them with
  [deploy/docker/docker-compose.yaml](../deploy/docker/docker-compose.yaml)
  after changing its `:latest` tags to `:v0.0.1`, then repeat the smoke
  tests of 14.3 (`INS`, `EDG`).
- Add a new `## [Unreleased]` section to the changelog, and set
  `core.Version` to the next pre-release (`v0.0.2-beta`).
