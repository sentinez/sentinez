

# Sentinez Guide

[![License](https://img.shields.io/badge/License-Apache%202.0-blue?logo=apache)](http://www.apache.org/licenses/LICENSE-2.0)
[![License-img](https://img.shields.io/badge/License-Creative%20Commons-blue)](https://creativecommons.org/licenses/by/4.0/)
![Mwgo](https://img.shields.io/badge/Made%20with-Go-blue?logo=go)

<img src="https://i.imgur.com/5Z8Qm77.png" alt="sentinez" width="300px">

---

### 🛡️ Sentinel of the Seventh Layer

> [!WARNING]
> Please keep in mind that ***Sentinez*** is still under active development
> and therefore full backward compatibility is not guaranteed before reaching v1.0.0.

#### Overview
**[SYSTEM ARCHITECTURE](./docs/i18n/en/system-architecture.md)**

----

**Main features** of *Sentinez* WAF as a service:
- Rule-based request filtering (based on OWASP CRS basics)
- Whitelist/blacklist configuration
- Provides a console interface for data visualization

How to build and run:  
**Requirements** before build and run:
- NodeJS 20.9.0
- Go 1.25.0
- Protocol Buffer
- Make (for running commands efficiently)

Clone source code with command:
```sh
git clone --recurse-submodules https://github.com/sentinez/sentinez.git $GOPATH/src/github.com/sentinez/sentinez
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
make apiserver.run
```

### License

Copyright (c) Sentinez Labs. All rights reserved.

Licensed under the [Apache 2.0](LICENSE) license.

### Logo

<p align="left">
  <img src="./assets/logo/sntz.png" width="40"/>
  <img src="./assets/logo/sntz-black.png" width="40"/>
  <img src="./assets/logo/sntz-light.png" width="40"/>
  <img src="./assets/logo/sntz-light2.png" width="40"/>
</p>

![Creative Commons License](https://i.creativecommons.org/l/by/4.0/88x31.png)  
**Sentinez** – photo by **Duc-Hung Ho**  
Licensed under [CC-BY-4.0](https://creativecommons.org/licenses/by/4.0/)

---
> **[© 2025 SENTINEZ](MADEINVIETNAM.md)** VN/CN/RU/IN/KR/US/JP/AU/FR/MY/NZ/ID/SG/TH