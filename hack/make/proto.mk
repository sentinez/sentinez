# Protobuf targets; included by the root Makefile, run from $(SZ_ROOT).

.PHONY: fmt.proto

fmt.proto:
	@cd $(SZ_ROOT)/api && buf format -w
