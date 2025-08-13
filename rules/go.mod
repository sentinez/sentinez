module github.com/sentinez/sentinez/rules

go 1.24.2

replace github.com/sentinez/sentinez/api => ../api

require (
	github.com/antlr4-go/antlr/v4 v4.13.1
	github.com/sentinez/sentinez/api v0.0.0
)

require (
	golang.org/x/exp v0.0.0-20240506185415-9bf2ced13842 // indirect
	google.golang.org/protobuf v1.36.5 // indirect
)
