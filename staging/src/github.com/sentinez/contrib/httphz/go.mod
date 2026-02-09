module github.com/sentinez/contrib/httphz

go 1.25.3

replace (
	github.com/sentinez/core => ../../core
	github.com/sentinez/sentinez/api => ../../../../../../api
	github.com/sentinez/shared => ../../shared
)

require (
	github.com/a-h/templ v0.3.960
	github.com/cloudwego/hertz v0.10.3
	github.com/gorilla/websocket v1.5.3
	github.com/hertz-contrib/http2 v0.1.8
	github.com/hertz-contrib/reverseproxy v1.0.6
	github.com/sentinez/core v0.0.0-00010101000000-000000000000
	github.com/sentinez/sentinez/api v0.0.0
	github.com/sentinez/shared v0.0.0-00010101000000-000000000000
)

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.10-20251209175733-2a1774d88802.1 // indirect
	github.com/bytedance/gopkg v0.1.3 // indirect
	github.com/bytedance/sonic v1.14.2 // indirect
	github.com/bytedance/sonic/loader v0.4.0 // indirect
	github.com/cloudwego/base64x v0.1.6 // indirect
	github.com/cloudwego/gopkg v0.1.7 // indirect
	github.com/cloudwego/netpoll v0.7.2 // indirect
	github.com/exaring/ja4plus v0.0.3 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/hertz-contrib/websocket v0.2.0 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/nyaruka/phonenumbers v1.6.7 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.2.0 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.1 // indirect
	golang.org/x/arch v0.23.0 // indirect
	golang.org/x/exp v0.0.0-20251209150349-8475f28825e9 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	google.golang.org/grpc v1.77.0 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)
