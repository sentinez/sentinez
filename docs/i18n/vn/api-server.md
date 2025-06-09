# SENTINEZ - API SERVER
Tài liệu kỹ thuật cho máy chủ API của dự án **Sentinez** đây đóng vai trò là 
backend cho toàn bộ hệ thống.

> [!NOTE]
> Ban đầu, tất cả các dịch vụ thuộc ```internal/core/*``` đều được nhúng trực
> tiếp vào ```apiserver```, và coi chúng như là một hệ thống monolithic truyền
> thống. Khi cần mở rộng hệ thống, ta có thể dễ dàng tách các dịch vụ trong 
> ```core``` ra thành 1 microservice với ```apiserver``` đóng vai trò là 
> API Gateway, gọi các dịch vụ trong ```core``` bằng giao thức gRPC

## 1. Đặc tả
*Sentinez* là dự án được triển khai hoàn toàn trên môi trường đám mây, việc 
thiết kế hệ thống Backend phải đáp ứng được những yêu cầu sau:
- Ứng dụng dễ dàng triển khai trên bất kì cloud plaform nào.
- Ứng dụng phải nhẹ và linh hoạt, thời gian phát triển nhanh để đáp ứng được nhu cầu cấp bách của thị trường.
- Ứng dụng phải chịu được tải lớn và dễ dàng mở rộng cả chiều rộng lẫn chiều sâu.

*API Server* như là trái tim của dự án. Server phải đủ sức mạnh để xử lý lượng lớn yêu cầu của người dùng.

## 2. Yêu cầu kỹ thuật
Những kỹ năng bắt buộc phải có
- Ngôn ngữ lập trình Go
- Protocol Buffer
- Kỹ năng về GRPC Gateway, GRPC Service
- PostgreSQL, Docker, REST API

## 3. Cấu trúc file & thư mục
API Server sẽ được hiện thực trên 2 thư mục chính.
```
cmd/apiserver/*
internal/apiserver/*
```

Ngoài ra, thư mục ```pkg``` chứa các gói phụ trợ cho dịch vụ này.

Chạy ứng dụng bằng câu lệnh sau, từ thư mục root của dự án ```$GOPATH/src/github.com/sentinez/sentinez/```:
```bash
go run cmd/apiserver/main.go
```

Ở thư mục chứa business logic chính ```internal/apiserver``` ta có các thành phần sau
- `handlers`: hiện thực các custom handler mà không cần thiết phải thông qua `internal/core`. vd: open API Swagger handler request
- `middleware`: setup các middleware http request cho apiserver
- `registrar`: hiện thực cách gọi để đăng kí dịch vụ bằng cả 2 cách. Nhúng trực tiếp service vào apiserver hoặc đăng kí endpoint microservice
- `apiserver.go`: hiện thực `sentinez.Server` interface để build ứng dụng bằng hàm `sentinez.Build()`
- `bootloader.go`: chạy các hàm `visit` để kích hoạt đăng kí dịch vụ thông qua gói `registrar` khi server khởi động
