# Targets for cmd/szapiserver; included by the root Makefile.
# Override the binary name with SENTINEZ_OUT and the image with TAG.
SZAPISERVER_DIR := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))

.PHONY: sz.apiserver.build sz.apiserver.run sz.apiserver.image.build

sz.apiserver.build: SENTINEZ_OUT ?= apiserver
sz.apiserver.build:
	@go build -ldflags="-s -w" -o $(SZAPISERVER_DIR)/bin/$(SENTINEZ_OUT) $(SZAPISERVER_DIR)
	@echo "[DONE]  senz: sz.apiserver ... ok"

sz.apiserver.run: SENTINEZ_OUT ?= apiserver
sz.apiserver.run: sz.apiserver.build
	@$(SZAPISERVER_DIR)/bin/$(SENTINEZ_OUT) --env_file=$(SZAPISERVER_DIR)/.env

sz.apiserver.image.build: TAG ?= sentinez/apiserver
sz.apiserver.image.build:
	@docker buildx build -f $(SZAPISERVER_DIR)/Dockerfile -t $(TAG):latest $(SZAPISERVER_DIR)/../..
