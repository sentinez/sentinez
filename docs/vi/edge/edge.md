# DMZ Gateway Lab Environment for XDP, WAF and Reverse Proxy Testing

## Overview

Tài liệu này mô tả môi trường lab dùng để phát triển và kiểm thử các thành phần bảo mật và xử lý lưu lượng mạng ở Layer 3–7, bao gồm:

* XDP/eBPF
* Web Application Firewall (WAF)
* Reverse Proxy
* Rate Limiting
* DDoS Mitigation

Mục tiêu là mô phỏng kiến trúc triển khai thực tế (production) trong đó Gateway được đặt trong vùng DMZ (Demilitarized Zone) và hoạt động như điểm tiếp nhận lưu lượng từ Internet trước khi chuyển tiếp tới Origin Server.

Toàn bộ môi trường được xây dựng bằng Linux Network Namespace nhằm tạo ra các network stack độc lập trên cùng một máy chủ, cho phép kiểm thử các luồng mạng thực tế mà không cần sử dụng máy ảo hoặc hạ tầng vật lý riêng biệt.

---

## Architecture

Môi trường bao gồm ba thành phần chính:

### Client Namespace

Mô phỏng người dùng hoặc hệ thống bên ngoài Internet.

Client gửi request tới Gateway thông qua interface public của DMZ.

### Gateway Namespace (DMZ)

Mô phỏng Reverse Proxy hoặc Edge Gateway trong môi trường production.

Gateway chịu trách nhiệm:

* Nhận lưu lượng từ Internet
* Thực thi XDP/eBPF
* Thực thi WAF
* Áp dụng Rate Limiting hoặc DDoS Protection
* Forward request tới Origin Server

Interface `veth0` đóng vai trò là public-facing network interface và là nơi chương trình XDP được attach.

### Host Network

Đóng vai trò uplink router cho Gateway.

Host thực hiện:

* IP Forwarding
* NAT
* Kết nối Gateway tới Internet thực

Origin Server không nằm trong môi trường lab mà là các website hoặc API thực trên Internet.

Ví dụ:

* https://google.com
* https://github.com
* https://httpbin.org
* Hệ thống backend thực tế của doanh nghiệp

---

## Network Topology

```text
                              Internet
                                  │
                                  ▼
                           Origin Server
                                  ▲
                                  │
                       NAT / Forwarding
                                  │
                                  ▼
+------------------------------------------------------+
|                    Host Linux                        |
|                                                      |
|   host-uplink (192.168.100.1/24)                     |
+-------------------------▲----------------------------+
                          │
                          │
                          ▼
+------------------------------------------------------+
|                 gateway-ns (DMZ)                     |
|                                                      |
|  veth0      : 203.0.113.1/24                         |
|  dmz-uplink : 192.168.100.2/24                       |
|                                                      |
|  Components:                                         |
|    - XDP/eBPF                                        |
|    - WAF                                             |
|    - Reverse Proxy                                   |
+-------------------------▲----------------------------+
                          │
                          │
                          ▼
+------------------------------------------------------+
|                    client-ns                         |
|                                                      |
|   veth-client : 203.0.113.2/24                       |
+------------------------------------------------------+
```

---

## Packet Flow

Một request từ client sẽ đi qua các thành phần theo thứ tự:

```text
Client
  │
  ▼
veth-client
  │
  ▼
veth0
  │
  ▼
XDP
  │
  ▼
WAF
  │
  ▼
Reverse Proxy
  │
  ▼
dmz-uplink
  │
  ▼
Host NAT
  │
  ▼
Internet
  │
  ▼
Origin Server
```

Trong quá trình xử lý, chương trình XDP được attach trên `veth0` sẽ nhìn thấy toàn bộ lưu lượng đi vào Gateway trước khi packet được chuyển lên Linux Networking Stack.

Ví dụ:

```text
Source IP      = 203.0.113.2
Destination IP = 203.0.113.1
```

Điều này cho phép kiểm thử chính xác các cơ chế:

* IP Filtering
* Geo Blocking
* SYN Flood Protection
* Rate Limiting
* DDoS Mitigation
* Early Packet Drop bằng XDP

---

## Why This Architecture

Mô hình này được lựa chọn vì mang lại nhiều đặc điểm tương đồng với môi trường production:

* Gateway được cô lập hoàn toàn khỏi host.
* Có public-facing interface riêng (`veth0`).
* Hỗ trợ attach XDP ở vị trí tương đương NIC public.
* Có đường outbound thực ra Internet.
* Không cần backend giả lập.
* Có thể kiểm thử đầy đủ luồng Client → Gateway → Origin.
* Có thể mở rộng thêm nhiều client namespace để mô phỏng lưu lượng lớn hoặc tấn công DDoS.

Kiến trúc này phù hợp cho việc phát triển và kiểm thử các hệ thống Edge Proxy, WAF, API Gateway hoặc DDoS Protection trước khi triển khai lên môi trường thực tế.
