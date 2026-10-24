# Targets for cmd/szgreeter; included by the root Makefile.
# Override the binary name with SENTINEZ_OUT and the image with TAG.
SZGREETER_DIR := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))

.PHONY: sz.greeter.build sz.greeter.run sz.greeter.image.build

sz.greeter.build: SENTINEZ_OUT ?= greeter
sz.greeter.build:
	@go build -ldflags="-s -w" -o $(SZGREETER_DIR)/v1/bin/$(SENTINEZ_OUT) $(SZGREETER_DIR)/v1
	@echo "[DONE]  senz: sz.greeter.v1 ... ok"

sz.greeter.run: SENTINEZ_OUT ?= greeter
sz.greeter.run: sz.greeter.build
	@$(SZGREETER_DIR)/v1/bin/$(SENTINEZ_OUT) --env_file=$(SZGREETER_DIR)/v1/.env

sz.greeter.image.build: TAG ?= sentinez/greeter
sz.greeter.image.build:
	@docker buildx build -f $(SZGREETER_DIR)/v1/Dockerfile -t $(TAG):latest $(SZGREETER_DIR)/../..
