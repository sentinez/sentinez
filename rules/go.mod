module github.com/sentinez/sentinez/rules

go 1.24.2

replace github.com/sentinez/sentinez/api => ../api

require (
	github.com/antlr4-go/antlr/v4 v4.13.1
	github.com/sentinez/sentinez/api v0.0.0
)

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.5-20250219170025-d39267d9df8f.1 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	golang.org/x/exp v0.0.0-20240506185415-9bf2ced13842 // indirect
	google.golang.org/protobuf v1.36.5 // indirect
)
