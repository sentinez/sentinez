

# Sentinéz /sen-ti-nɛz/

[![License](https://img.shields.io/badge/License-Apache%202.0-blue?logo=apache)](http://www.apache.org/licenses/LICENSE-2.0)
[![License-img](https://img.shields.io/badge/License-Creative%20Commons-blue)](https://creativecommons.org/licenses/by/4.0/)
![Mwgo](https://img.shields.io/badge/Made%20with-Go-blue?logo=go)

![img](./_logo/senzz.png)

---
**Sentinez** *stylized as* **Sentinéz**

### 🛡️ Sentinel - Edge reverse proxy with many antioxidants

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
- Go 1.26+
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
make apiserver.run
```

**Run** edge proxy
```sh
make edge.run
```

### License

Copyright (c) Sentinéz Labs. All rights reserved.

Licensed under the [Apache 2.0](LICENSE) license.

### Logo

<p align="left">
  <img src="./_logo/sntz.png" width="40"/>
  <img src="./_logo/sntz-black.png" width="40"/>
  <img src="./_logo/sntz-light.png" width="40"/>
  <img src="./_logo/sntz-light2.png" width="40"/>
</p>

![Creative Commons License](https://i.creativecommons.org/l/by/4.0/88x31.png)  
**Sentinéz** – photo by **Duc-Hung Ho**  
Licensed under [CC-BY-4.0](https://creativecommons.org/licenses/by/4.0/)

---
> **[© 2025-2026 SENTINEZ](COO.md)** VN/CN/RU/IN/KR/US/JP/AU/FR/MY/NZ/ID/SG/TH/UK/EU