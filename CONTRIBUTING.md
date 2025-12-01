# Contributors' Guide
This guide will assist contributors to the `sentinez/sentinez` repository.

## Prerequisites

### Applications used in this project

- Backend Go & Protobuf: GoLand
- Frontend: Visual Studio Code or any preferred IDE

### Go Coding Style

We follow the Uber Go Style Guide in this project. Please ensure you are familiar with the Go programming style as per the guide.

> See more at: https://github.com/uber-go/guide

Ensure that:

- Go files comply with the standard Go formatting and style.
- Protobuf files comply with the format and use tabs for indentation.
- Files do not contain trailing whitespaces and must end with a single newline character.

Use the following command to ensure your code meets expectations:

```
make lint
```

### Conventional Commits

All pull requests to the main branch must adhere to Conventional Commits. Otherwise, they will not be accepted.

> See more at: https://www.conventionalcommits.org/en/v1.0.0/

### Applying License Header to New Files

Ensure that you have added the license header to each file you create, including `.go`, `makefile`, `.sh`, `Dockerfile`, etc.

```
Copyright 2025 Duc-Hung Ho.

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

### Sign Your Work

The sign-off is a simple line at the end of the explanation for the patch. Your signature certifies that you wrote the patch or otherwise have the right to pass it on as an open-source patch. The rules are simple: if you can certify the below (from https://developercertificate.org/):

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
Signed-off-by: Duc Hung Ho <hunghd@email.com>
```

If you set your `user.name` and `user.email` git configs, you can sign your commit automatically with `git commit -s`.

## Naming Packages

**Where to Put Packages**
- `/api`: define proto files for all service mesh.
- `/cmd`: main application apis, edge app endpoints.
- `/deploy`: contains scripts and config for deployment.
- `/docs`: documentation.
- `/hack`: scripts used by developers.
- `/internal`: internal packages, main business logic, not exported.
- `/pkg`: common packages, shared with external modules.
- `/staging`: external repository staging area
- `/ui`: main application web ui based on monorepo
