# Tài liệu kỹ thuật Sentinéz

[English](../README.md) | **Tiếng Việt**

Bộ tài liệu mô tả mã nguồn Sentinéz — nền tảng bảo mật viết bằng Go gồm
edge proxy có WAF (OWASP CRS qua Coraza), rule engine tuỳ biến (SecRule),
rate limiting, CDN cache, dataplane eBPF/XDP và web console.

Tài liệu được viết từ việc đọc trực tiếp source code tại commit `47e5f876`
(nhánh `main`). Khi code thay đổi, hãy cập nhật file tương ứng.

## Mục lục

| # | Tài liệu | Nội dung |
|---|---|---|
| 1 | [Tổng quan kiến trúc](01-architecture.md) | Các service, module, luồng dữ liệu tổng thể |
| 2 | [Bắt đầu phát triển](02-getting-started.md) | Yêu cầu, build, chạy, biến môi trường, flag |
| 3 | [Edge proxy](03-edge.md) | Pipeline xử lý request, cấu hình `proxy.yaml`, memory store, cluster |
| 4 | [Rule engine & WAF](04-rule-engine.md) | SecRule expression, toán tử, CDN rule, Coraza/CRS, rule parser |
| 5 | [Dataplane eBPF](05-dataplane-ebpf.md) | Chương trình XDP, BPF map, netns dev |
| 6 | [Gateway: API server & realtime](06-gateway.md) | grpc-gateway, middleware, Swagger, WebSocket, service discovery |
| 7 | [Control plane](07-controlplane.md) | IAM, tenant, security, analytic, greeter, centraldata; lớp DB; xác thực |
| 8 | [API & Protobuf](08-api-protobuf.md) | Cấu trúc proto, custom option, codegen, bảng REST endpoint |
| 9 | [Framework & thư viện dùng chung](09-core-shared.md) | `core/runner`, `core/http`, `zlog`, `config`, `errorx`, ... |
| 10 | [Web console (UI)](10-ui.md) | Workspace pnpm/turbo, app console, gọi API |
| 11 | [Triển khai, CI & phát hành module](11-deploy-ci.md) | Docker, monitoring, GitHub Actions, đồng bộ staging |
| 12 | [Các vấn đề đã biết](12-known-issues.md) | Lỗi, phần chưa hoàn thiện và rủi ro phát hiện khi đọc code |

## Quy ước đọc tài liệu

- Đường dẫn file là tương đối so với gốc repo, ví dụ
  [internal/dmz/edge/http/init.go](../../internal/dmz/edge/http/init.go).
- Module trong `staging/src/github.com/sentinez/<tên>` được gọi tắt theo
  import path: `core`, `shared`, `controlplane`, `contrib/httphz`, `bpf`,
  `tools`.
- "SecRule" là rule tuỳ biến của người dùng (tên cũ `RuleBased` đã bỏ);
  "rulesets"/"WAF" là bộ OWASP CRS chạy trên Coraza.
