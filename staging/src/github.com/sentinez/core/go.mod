module github.com/sentinez/core

go 1.25.3

replace (
	github.com/sentinez/sentinez/api => ../../../../../api
	github.com/sentinez/shared => ../shared
)

require (
	github.com/a-h/templ v0.3.960
	github.com/antlr4-go/antlr/v4 v4.13.1
	github.com/corazawaf/coraza/v3 v3.3.3
	github.com/google/uuid v1.6.0
	github.com/gorilla/websocket v1.5.3
	github.com/sentinez/sentinez/api v0.0.0
	github.com/sentinez/shared v0.0.0-00010101000000-000000000000
	go.uber.org/fx v1.24.0
	google.golang.org/grpc v1.77.0
	google.golang.org/protobuf v1.36.10
)

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.10-20251209175733-2a1774d88802.1 // indirect
	github.com/cloudresty/ulid v1.2.1 // indirect
	github.com/corazawaf/libinjection-go v0.2.3 // indirect
	github.com/magefile/mage v1.15.1-0.20241126214340-bdc92f694516 // indirect
	github.com/matoous/go-nanoid/v2 v2.1.0 // indirect
	github.com/petar-dambovaliev/aho-corasick v0.0.0-20250424160509-463d218d4745 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	github.com/rs/xid v1.6.0 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.2.0 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/valllabh/ocsf-schema-golang v1.0.3 // indirect
	go.uber.org/dig v1.19.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.1 // indirect
	golang.org/x/exp v0.0.0-20251209150349-8475f28825e9 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	rsc.io/binaryregexp v0.2.0 // indirect
)
