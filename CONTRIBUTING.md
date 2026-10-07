# Contributors' Guide

This guide will assist contributors to the `sentinez/sentinez` repository.

## Prerequisites

### Toolchain

- Go 1.27+
- Node 26 (see `engines` in [ui/package.json](ui/package.json)) and pnpm
- [buf](https://buf.build) for protobuf
- `make`, `docker` (for images and `make compose.up`)
- [golangci-lint](https://golangci-lint.run) (`hack/install_golint.sh`)
- Other dev tools: `hack/install_dev_tools.sh` (migrate, wire, mockery,
  templ, ...)

### Getting the Source

The repository uses git submodules (coreruleset, googleapis, grpc-gateway,
protovalidate, and the libbpf/xdp-tools submodules of `staging/.../bpf`):

```sh
git clone --recurse-submodules https://github.com/sentinez/sentinez.git
# or, for an existing clone
git submodule update --init --recursive
```

### Editors

- Backend Go & Protobuf: GoLand or Visual Studio Code
- Frontend: Visual Studio Code or any preferred IDE

## Repository Layout

This is a **multi-module monorepo**. The root [go.mod](go.mod) uses
`replace` directives to wire in the other modules; there is no `go.work`.

- `/api`: separate module, protobuf sources under `proto/sentinez/`
  (`apps`, `dmz`, `gateway`, `types`) and generated code. Shared messages
  live under `types/` (`v1`, `setting`, `rule`, `secrule`, `coreruleset`,
  `cdn`, `net`, `net/http`).
- `/cmd`: service entrypoints (`szapiserver`, `szrealtime`, `szedge`,
  `szdataplane`, `szgreeter`, `szcentraldata`).
- `/deploy`: deployment scripts and configs (caddy, nginx, docker,
  monitoring, CRS rules).
- `/docs`: documentation.
- `/hack`: scripts used by developers (lint, proto lint, timescale, mkcert,
  rule parsing, ...).
- `/internal`: private code of the root module, not exported.
- `/pkg`: shared packages of the root module, including service wiring under
  `pkg/apps`.
- `/staging/src/github.com/sentinez`: separately-versioned modules published
  to their own repositories (`core`, `shared`, `controlplane`, `bpf`,
  `contrib/httphz`, `tools`).
- `/ui`: pnpm + turbo workspace for the web console and shared UI packages.
- `/_submodules`: third-party git submodules.

Binaries live in `cmd/sz<name>`. Domain protos live under
`api/proto/sentinez/apps/<domain>/v1`; shared proto types under
`api/proto/sentinez/types/<name>/v1`. WAF rules are called `SecRule`
(the old `RuleBased` name is retired).

## Building and Testing

```sh
make                      # build all default services into cmd/<svc>/bin/
make sz.<svc>.build       # apiserver, realtime, edge, dataplane, greeter,
                          # centraldata
make sz.<svc>.run         # build + run
make sz.<svc>.image.build # docker image
make compose.up           # start the local stack
make test.cover           # go test ./... -cover (root module only)
```

Because the modules are separate, `go test ./...` at the root does **not**
cover the `staging/` modules. Run tests inside each module you touch:

```sh
cd staging/src/github.com/sentinez/shared && go test ./...
```

After changing dependencies in a staged module, run `go mod tidy` in that
module and in the root.

Generated code:

- Protobuf: `cd api && buf generate`, format with `make fmt.proto`.
- Mocks: `mockery`.
- Templates: `templ generate`.

Commit generated files together with their sources.

## Go Coding Style

We follow the [Uber Go Style Guide](https://github.com/uber-go/guide). The
project-specific rules are in
[.agent/skills/go-style-guide/SKILL.md](.agent/skills/go-style-guide/SKILL.md);
please read it before writing or reviewing Go code.

The following are enforced by [.golangci.yaml](.golangci.yaml):

- `gofmt` + `goimports` formatting; imports in two groups (stdlib, then
  everything else).
- Lines of at most 80 columns (tab = 4), functions of at most 35 lines.
- `errcheck`, `govet`, `staticcheck`, `misspell`, `revive`.

Also ensure that:

- Protobuf files pass `buf lint` and `buf format`.
- Files do not contain trailing whitespace and end with a single newline.

Run the linters before opening a pull request:

```sh
make lint          # all Go modules + protobuf
make lint.core     # or a single module: lint.shared, lint.controlplane,
                   # lint.contrib.httphz, lint.tools, lint.proto
```

Tests are table-driven with subtests and use testify `require`/`assert`.

## Conventional Commits

All pull requests to the main branch must adhere to
[Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/)
(`feat:`, `fix:`, `refactor:`, `chore:`, ...). Otherwise, they will not be
accepted. Commits on main carry the PR number as a suffix, e.g.
`refactor: bring dmz region back (#295)`.

CI runs lint (`hack/golint.sh`, `hack/protolint.sh`) and tests on every pull
request; both must pass.

## Applying License Header to New Files

Ensure that you have added the license header to each file you create,
including `.go`, `Makefile`, `.sh`, `Dockerfile`, etc.

```
Copyright 2025-2026 Duc-Hung Ho.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```

## Sign Your Work

The sign-off is a simple line at the end of the explanation for the patch.
Your signature certifies that you wrote the patch or otherwise have the right
to pass it on as an open-source patch. The rules are simple: if you can
certify the below (from https://developercertificate.org/):

```
Developer Certificate of Origin
Version 1.1

Copyright (C) 2004, 2006 The Linux Foundation and its contributors.
660 York Street, Suite 102,
San Francisco, CA 94110 USA

Everyone is permitted to copy and distribute verbatim copies of this
license document, but changing it is not allowed.

Developer's Certificate of Origin 1.1

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I
    have the right to submit it under the open source license
    indicated in the file; or

(b) The contribution is based upon previous work that, to the best
    of my knowledge, is covered under an appropriate open source
    license and I have the right under that license to submit that
    work with modifications, whether created in whole or in part
    by me, under the same open source license (unless I am
    permitted to submit under a different license), as indicated
    in the file; or

(c) The contribution was provided directly to me by some other
    person who certified (a), (b) or (c) and I have not modified
    it.

(d) I understand and agree that this project and the contribution
    are public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.
```

Then you just add a line to every git commit message:

```
Signed-off-by: Your Name <your.email@example.com>
```

If you set your `user.name` and `user.email` git configs, you can sign your
commit automatically with `git commit -s`.
