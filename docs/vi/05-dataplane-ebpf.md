# 5. Dataplane eBPF (`szdataplane`)

Dataplane gắn một chương trình XDP vào interface mạng để xử lý gói tin ở tầng
L3 trước khi tới edge.

| Đường dẫn | Vai trò |
|---|---|
| [bpf/sentinez/sz.c](../../staging/src/github.com/sentinez/bpf/sentinez/sz.c) | Chương trình XDP `sz_main` |
| [bpf/sentinez/sz_monitor.h](../../staging/src/github.com/sentinez/bpf/sentinez/sz_monitor.h) | Đếm byte theo IP nguồn |
| [bpf/sentinez/sz_security.h](../../staging/src/github.com/sentinez/bpf/sentinez/sz_security.h) | Chặn IP/CIDR bằng LPM trie |
| `bpf/sentinez/sz_bpfe{l,b}.{go,o}` | Code Go + object sinh bởi `bpf2go` |
| [internal/apps/dataplane/](../../internal/apps/dataplane/) | `Server` (Start/Stop) |
| [internal/apps/dataplane/driver/](../../internal/apps/dataplane/driver/) | `Context`: nạp object, attach XDP, đọc ghi map (`LookupBandwidth`, `BlockCIDR`) |
| [hack/net/dev/Makefile](../../hack/net/dev/Makefile) | Dựng network namespace để thử nghiệm |

## 5.1 Chương trình XDP

```c
SEC("xdp") int sz_main(struct xdp_md *ctx) {
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
cd staging/src/github.com/sentinez/bpf/sentinez && go generate .   # bpf2go Sz ./sz.c
```

Header libbpf/xdp-tools lấy từ submodule trong `bpf/_submodules`.

## 5.2 Phía Go

`driver.NewContext() (*Context, error)`: `rlimit.RemoveMemlock()` →
`LoadSzObjects` → giữ `*SzObjects`. `AttachXDP(ifIndex)` gắn
`SzMain` qua `link.AttachXDP`. `Close` gỡ link rồi mới dỡ object. Lỗi có
tiền tố `driver:`. Package không giữ trạng thái global: `Context` thuộc về
nơi tạo ra nó.

`dataplane.Server.Start(VETH0)`: lấy interface `veth0`, nạp một
`driver.Context` và attach XDP (attach lỗi thì đóng lại ngay), rồi phục vụ
gRPC (`DataPlaneService` mới có `Status` trong proto, chưa đăng ký handler).
Server giữ `Context`; `Stop` tắt gRPC rồi đóng nó, và trả về cả hai lỗi.

Các method của `driver.Context` trên map:

| Method | Mô tả |
|---|---|
| `LookupBandwidth(ip) (uint64, error)` | Cộng giá trị per-CPU của `ip_bandwidth` cho IPv4 (mỗi CPU khả dụng một giá trị, `ebpf.PossibleCPU`); 0 khi địa chỉ chưa có |
| `BlockCIDR(cidr) error` | Thêm khoá LPM `{prefixlen, ip mạng}` vào `blocklist` (chỉ IPv4) |

Khoá giữ IP theo thứ tự byte mạng trong bộ nhớ, giống `saddr` thô mà XDP
dùng (`binary.NativeEndian` trên 4 byte địa chỉ).

Lưu ý: map không được pin, nên các method này chỉ dùng được trên `Context`
của tiến trình dataplane. Hiện chưa có nơi nào gọi chúng.

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
