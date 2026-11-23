# github.com/sentinez/integration

> ⚠️ **This is an automatically published [staged repository](https://github.com/sentinez/integration) for Sentinez**.   
> Contributions, including issues and pull requests, should be made to the main Sentinez repository: [https://github.com/sentinez/sentinez](https://github.com/sentinez/sentinez).  
> This repository is read-only for importing, and not used for direct contributions.

=======================================================================

Integrations of Sentinez with external systems and protocols, such as the
Model Context Protocol (MCP).

Each integration is a self-contained package that depends on the other
Sentinez modules (`core`, `pkg`) and is never imported by them, so a
service links only the integrations it uses.

## Install

```sh
go get github.com/sentinez/integration
```

## License

Apache License 2.0, see [LICENSE](LICENSE).
