# Về Sentinéz

> **Author:** Hồ Đức Hưng  
> **Organization:** Sentinez  
> **Version:** v1.0  
> **Date:** 2025-12-06

***Sentinez***, viết cách điệu ***Sentinéz*** là một chuỗi phần mềm cho cơ sở hạ
tầng đám mây, được viết bằng các ngôn ngữ như Go, Rust, Java.

## Lịch sử

Được sáng lập bởi @kyerans (Duc Hung Ho) vào tháng 1/2025 khi đang học tập tại
Trường đại học Bách Khoa, Đại học quốc gia Hồ Chí Minh (HCMUT-VNUHCM) và làm việc
tại phòng điện toán đám mây, Công ty cổ phần giải pháp thanh toán Việt Nam (VNPAY)

## Về Sentinéz Stack

Chuỗi ứng dụng này gồm có:
- **Sentinez Edge** máy chủ proxy ngược, nơi tiếp nhận các yêu cầu và phản hồi 
ở lớp biên, trực tiếp của người dùng, tích hợp các tính năng bảo mật cho địa chỉ
web đồng thời bảo vệ kết nối bằng TLS
- **API Server** cung cấp các api tương tác từ người dùng lên hệ thống của sentinez
- **Realtime** cho phép streaming data thông qua websocket từ hệ thống tới người dùng
- **Catinat** giải pháp cung cấp dịch vụ theo dõi hệ thống, được viết bằng ngôn
ngữ Java mạnh mẽ.

Ngoài ra, Sentinez còn phát triển 1 số công cụ để bảo vệ hạ tầng web ở mức độ
sâu hơn
- **Etheral**, viết cách điệu ***Étheral***, là một bản proxy ngược được viết
bằng ngôn ngữ Rust mạnh mẽ và an toàn cho bộ nhớ
- **Quadrum** cung cấp các cơ chế mạnh mẽ để theo dõi và thực hiện các giải pháp bảo
vệ ở sâu L3/L4, trong khi hệ thống Sentinez bảo vệ người dùng chủ yếu ở L7 trong
mô hình OSI

## Tính năng

Sentinez cung cấp cho người dùng các chứng năng bảo mật website tiêu biểu
- Rule-based do người dùng tạo, dùng để cảnh báo, chặn các request trực tiếp
- WAF cung cấp bởi OWASP corerulesets, 1 tập hợp rules dùng để lọc request một
cách hiệu quả
- Các rule limit để chống lại các cuộc tấn công DDOS
- Lọc các package và chặn các IP độc hại ngay ở L3/L4
- Cung cấp các monitor rule giúp cảnh báo sớm các mối nguy hiểm
- Bảo mật website và routing địa chỉ website bằng TLS/SSL

## Cloud Native

Tất cả các ứng dụng được biên dịch và ảo hóa bằng Docker, với kích thước của mỗi
image siêu nhẹ, cho phép các lập trình viên có thể tối ưu việc lưu trữ và triển
khai lên các hạ tầng đám mây nhanh chóng bằng Docker, Docker compose hay Kubernetes

## Hiệu suất cao

Đa phần các ứng dụng được viết và biên dịch bằng Go, ngôn ngữ native mạnh mẽ từ
Google, sử dụng các công nghệ tiên tiến nhất của thời điểm hiện tại (ProtocolBuffer,
gRPC Ecosystem) mang đến hiệu suất và sự mở rộng nhanh chóng kèm theo là độ trễ
được đánh giá tương đối thấp, Sentinez cung cấp yêu cầu triển khai nhanh chóng
trên bất kì cụm máy chủ nào và vẫn giữ được trạng thái ổn định.