<img src="./_logo/sentinez/3/3.png" width="600"/>

**Sentinez** *stylized as* **Sentinéz** /sen-ti-nɛz/

[![License](https://img.shields.io/badge/License-Apache%202.0-blue?logo=apache)](http://www.apache.org/licenses/LICENSE-2.0)
[![License-img](https://img.shields.io/badge/License-Creative%20Commons-blue)](https://creativecommons.org/licenses/by/4.0/)
![Mwgo](https://img.shields.io/badge/Made%20with-Go-blue?logo=go)

### 🛡️ Sentinéz - Intelligent security platform with many antioxidants

> [!WARNING]
> Please keep in mind that ***Sentinéz*** is still under active development
> and therefore full backward compatibility is not guaranteed before reaching v1.0.0.

----

**Main features** of *Sentinéz*
- Rule-based request filtering (based on OWASP CRS basics)
- Rate limiter configuration
- Provides a console interface for data visualization

How to build and run:  
**Requirements** before build and run:
- NodeJS 20.9.0
- Go 1.27+
- Protocol Buffer
- Make (for running commands efficiently)

Clone source code with command:
```sh
git clone https://github.com/sentinez/sentinez.git $GOPATH/src/github.com/sentinez/sentinez
``` 

If you have already cloned the repo, you can initialize the submodule with:

```sh
git submodule update --init --recursive
```

> [!NOTE]
> You can view mirror project from:
> https://gitlab.com/sentinez/sentinez.git

**Run** the project with `apiserver`
```sh
make sz.apiserver.run
```

**Run** edge proxy
```sh
make sz.edge.run
```

### Benchmarks

The cost of the edge request path: the chain plugins (`internal/handlers`)
and the `net/http` context they run on (`pkg/network/httpx`). Refresh the
table with:

```sh
make bench                # or BENCH_COUNT=10 make bench
```

<!-- bench:start -->

_go1.27.1, linux/amd64, 13th Gen Intel(R) Core(TM) i7-1355U; median of 6 runs, 2026-11-20._

| Package | Benchmark | ns/op | B/op | allocs/op |
|---|---|--:|--:|--:|
| `internal/handlers` | StageDenied | 400 | 0 | 0 |
| `internal/handlers/cdn` | CacheHit | 1054 | 240 | 18 |
| `internal/handlers/cdn` | CacheMissUncacheable | 1044 | 384 | 28 |
| `internal/handlers/logging` | NewEvent | 1440 | 1072 | 3 |
| `internal/handlers/logging` | NewEgressEvent | 348 | 48 | 2 |
| `internal/handlers/secure` | SecLimiterAllowed | 179 | 0 | 0 |
| `internal/handlers/static` | StaticAsset | 90.3 | 16 | 1 |
| `internal/handlers/wrap/script` | HooksOneStage | 1520 | 560 | 9 |
| `pkg/network/httpx` | ContextLifecycle | 617 | 80 | 4 |
| `pkg/network/httpx` | ContextGetters | 135 | 48 | 1 |
| `pkg/network/httpx` | ContextVisitRequestHeaders | 119 | 0 | 0 |
| `pkg/network/httpx` | ContextSetResponseHeader | 72.2 | 16 | 1 |

<!-- bench:end -->

### License

Copyright (c) Sentinéz Labs. All rights reserved.

Licensed under the [Apache 2.0](LICENSE) license.

### Logo

<p align="left">
  <img src="./_logo/sz.png" width="40"/>
  <img src="./_logo/sz-black.png" width="40"/>
  <img src="./_logo/sz-light.png" width="40"/>
  <img src="./_logo/sz-light2.png" width="40"/>
</p>

![Creative Commons License](https://i.creativecommons.org/l/by/4.0/88x31.png)  
**Sentinéz** – photo by **Duc-Hung Ho**  
[![HCMUT](./_logo/hcmut-badge.svg)](https://hcmut.edu.vn)

Licensed under [CC-BY-4.0](https://creativecommons.org/licenses/by/4.0/)

---
> **[© 2025-2026 SENTINEZ](COO.md)** VN/CN/RU/IN/KR/US/JP/AU/FR/MY/NZ/ID/SG/TH/UK/EU/DK
