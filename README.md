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

_go1.27.1, linux/amd64, 13th Gen Intel(R) Core(TM) i7-1355U; median of 6 runs, 2026-11-21._

| Package | Benchmark | ns/op | B/op | allocs/op |
|---|---|--:|--:|--:|
| `internal/handlers` | StageDenied | 660 | 0 | 0 |
| `internal/handlers/cdn` | CacheHit | 1671 | 240 | 18 |
| `internal/handlers/cdn` | CacheMissUncacheable | 1826 | 384 | 28 |
| `internal/handlers/logging` | NewEvent | 1658 | 1072 | 3 |
| `internal/handlers/logging` | NewEgressEvent | 537 | 48 | 2 |
| `internal/handlers/secure` | SecLimiterAllowed | 302 | 0 | 0 |
| `internal/handlers/secure` | SecLimiterExceeded | 9042 | 11024 | 36 |
| `internal/handlers/secure` | SecLimiterNoRules | 23.3 | 0 | 0 |
| `internal/handlers/secure` | SecRuleNoRules | 27.6 | 0 | 0 |
| `internal/handlers/secure` | SecRuleNoMatch | 87.0 | 0 | 0 |
| `internal/handlers/secure` | SecRuleBlock | 8356 | 10968 | 34 |
| `internal/handlers/secure` | SecRuleActions | 1197 | 80 | 3 |
| `internal/handlers/secure` | SecRuleChallenge | 27232 | 55560 | 41 |
| `internal/handlers/secure` | DomainBased | 108 | 0 | 0 |
| `internal/handlers/secure` | Misdirected | 32.7 | 0 | 0 |
| `internal/handlers/secure` | SecRuleset/clean | 656218 | 185990 | 3155 |
| `internal/handlers/secure` | SecRuleset/attack | 928707 | 166314 | 3236 |
| `internal/handlers/secure` | SecRulesetNoWAF | 5820 | 6961 | 27 |
| `internal/handlers/secure` | Evaluate | 60.3 | 0 | 0 |
| `internal/handlers/secure` | ApplyActions | 320 | 24 | 2 |
| `internal/handlers/secure` | Redirect | 402 | 83 | 5 |
| `internal/handlers/secure` | NewSecRuleEvent | 1018 | 96 | 3 |
| `internal/handlers/static` | StaticAsset | 134 | 16 | 1 |
| `internal/handlers/wrap/script` | HooksOneStage | 1938 | 560 | 9 |
| `pkg/network/httpx` | ContextLifecycle | 936 | 80 | 4 |
| `pkg/network/httpx` | ContextGetters | 160 | 48 | 1 |
| `pkg/network/httpx` | ContextVisitRequestHeaders | 191 | 0 | 0 |
| `pkg/network/httpx` | ContextSetResponseHeader | 95.0 | 16 | 1 |

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
