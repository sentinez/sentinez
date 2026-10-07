# 5. Dataplane eBPF (`szdataplane`)

Dataplane gắn một chương trình XDP vào interface mạng để xử lý gói tin ở tầng
L3 trước khi tới edge.

| Đường dẫn | Vai trò |
|---|---|
| [bpf/sentinez/senz.c](../../staging/src/github.com/sentinez/bpf/sentinez/senz.c) | Chương trình XDP `senz_main` |
| [bpf/sentinez/senz_monitor.h](../../staging/src/github.com/sentinez/bpf/sentinez/senz_monitor.h) | Đếm byte theo IP nguồn |
| [bpf/sentinez/senz_security.h](../../staging/src/github.com/sentinez/bpf/sentinez/senz_security.h) | Chặn IP/CIDR bằng LPM trie |
| `bpf/sentinez/senz_bpfe{l,b}.{go,o}` | Code Go + object sinh bởi `bpf2go` |
| [internal/dmz/dataplane/](../../internal/dmz/dataplane/) | `Server` (Start/Stop) và `driver` (nạp object, attach XDP) |
| [internal/bpf/](../../internal/bpf/) | API Go thao tác map: `LookupBandwidth`, `BlockCIDR` |
| [hack/net/dev/Makefile](../../hack/net/dev/Makefile) | Dựng network namespace để thử nghiệm |

## 5.1 Chương trình XDP

```c
SEC("xdp") int senz_main(struct xdp_md *ctx) {
    action = bandwidth_handler(ctx);    // luôn XDP_PASS
    if (action != XDP_PASS) return action;
    return security_rule_handler(ctx);  // XDP_DROP nếu IP nằm trong blocklist
}
```

| Map | Kiểu | Key → Value | Dùng bởi |
|---|---|---|---|
| `ip_bandwidth` | `LRU_PERCPU_HASH`, 10240 | `__u32 saddr` → `__u64 bytes` | `bandwidth_handler` cộng dồn `data_end - data` |
| `blocklist` | `LPM_TRIE`, 1024, `NO_PREALLOC` | `{prefixlen, ip}` → `__u8` | `security_rule_handler` tra với prefixlen 32 |

Chỉ xử lý IPv4 (`ETH_P_IP`); gói khác được `XDP_PASS`. Trong
`security_rule_handler`, gói ngắn hơn header Ethernet/IP bị `XDP_ABORTED`.
Macro `debug()` dùng `bpf_printk` (xem bằng
`cat /sys/kernel/tracing/trace_pipe`; Makefile netns có target `tracing`).

Biên dịch lại:

```sh
cd staging/src/github.com/sentinez/bpf/sentinez && go generate .   # bpf2go Senz ./senz.c
```

Header libbpf/xdp-tools lấy từ submodule trong `bpf/_submodules`.

## 5.2 Phía Go

`driver.NewContext()` (singleton): `rlimit.RemoveMemlock()` →
`LoadSenzObjects` → giữ `*SenzObjects`. `AttachXDP(ifIndex)` gắn
`SenzMain` qua `link.AttachXDP`. `driver.Exec(fn)` cho phép thao tác map khi
context đã được khởi tạo.

`dataplane.Server.Start(VETH0)`: lấy interface `veth0`, attach XDP rồi phục vụ
gRPC (`DataPlaneService` mới có `Status` trong proto, chưa đăng ký handler).
`Stop` đóng object/link và `GracefulStop` gRPC.

API trong `internal/bpf`:

| Hàm | Mô tả |
|---|---|
| `LookupBandwidth(ip) (uint64, error)` | Cộng giá trị per-CPU của `ip_bandwidth` cho IPv4 |
| `BlockCIDR(cidr) error` | Thêm khoá LPM `{prefixlen, ip}` vào `blocklist` (chỉ IPv4) |

Lưu ý: `driver` là trạng thái **trong tiến trình**. Edge gọi
`LookupBandwidth` trong node `LOG`, nhưng edge không nạp object eBPF (dataplane
là tiến trình khác, map không được pin) nên luôn trả lỗi/0. `BlockCIDR` hiện
chưa có nơi gọi. Xem [12-known-issues.md](12-known-issues.md).

## 5.3 Môi trường netns để phát triển

`hack/net/dev/Makefile` (được `include` từ Makefile gốc):

```
 netns "client"                 netns "gateway"                    host
 veth-client 203.0.113.2/24 ◄──► veth0 203.0.113.1/24  (XDP)
                                 dmz-uplink 192.168.100.2/24 ◄──► host-uplink 192.168.100.1/24 ──NAT──► Internet
```

| Target | Tác dụng |
|---|---|
| `setup` | `clean` + `create`: tạo netns, veth, IP, route mặc định, bật `ip_forward`, MASQUERADE |
| `shell-client` / `shell-gateway` | Mở shell trong netns |
| `test-public` | Từ client `curl http://203.0.113.1:7443` |
| `test-internet` | Từ gateway `curl -I https://google.com` |
| `status`, `clean`, `curl`, `tracing`, `tracing.debug` | Xem trạng thái, dọn dẹp, test, đọc trace eBPF |

Chạy dataplane trong netns gateway: `make sz.dataplane.run` (dùng `sudo ip
netns exec gateway`).
