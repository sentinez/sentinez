# Go and protobuf linting; included by the root Makefile, run from
# $(SZ_ROOT). Each staged module is linted in its own directory.
SZ_STAGING := $(SZ_ROOT)/staging/src/github.com/sentinez

.PHONY: lint lint.go lint.proto lint.tools lint.core lint.pkg \
	lint.controlplane lint.contrib.httphz

lint: lint.go lint.proto lint.core lint.pkg lint.controlplane \
	lint.contrib.httphz lint.tools

lint.go:
	@echo "[LINT] sentinez is linting ..."
	@cd $(SZ_ROOT) && golangci-lint run

lint.proto:
	@echo "[LINT] api is linting ..."
	@cd $(SZ_ROOT)/api && buf lint
	@cd $(SZ_ROOT)/api && golangci-lint run

lint.tools:
	@cd $(SZ_STAGING)/tools && golangci-lint run

lint.core:
	@echo "[LINT] core is linting ..."
	@cd $(SZ_STAGING)/core && golangci-lint run

lint.pkg:
	@echo "[LINT] pkg is linting ..."
	@cd $(SZ_STAGING)/pkg && golangci-lint run

lint.controlplane:
	@echo "[LINT] controlplane is linting ..."
	@cd $(SZ_STAGING)/controlplane && golangci-lint run

lint.contrib.httphz:
	@echo "[LINT] httphz is linting ..."
	@cd $(SZ_STAGING)/contrib/httphz && golangci-lint run
