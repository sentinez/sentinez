module github.com/sentinez/shared

go 1.25.0

replace github.com/sentinez/sentinez/api => ../../../../../api

require (
	github.com/exaring/ja4plus v0.0.2
	github.com/google/uuid v1.6.0
	github.com/matoous/go-nanoid/v2 v2.1.0
	github.com/sentinez/sentinez/api v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.27.0
	google.golang.org/protobuf v1.36.10
)

require (
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	go.uber.org/multierr v1.10.0 // indirect
)
