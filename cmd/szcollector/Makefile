# Targets for cmd/szcollector; included by the root Makefile.
# Override the binary name with SENTINEZ_OUT and the image with TAG.
SZCOLLECTOR_DIR := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))

.PHONY: sz.collector.build sz.collector.run sz.collector.image.build

sz.collector.build: SENTINEZ_OUT ?= collector
sz.collector.build:
	@go build -ldflags="-s -w" -o $(SZCOLLECTOR_DIR)/v1/bin/$(SENTINEZ_OUT) $(SZCOLLECTOR_DIR)/v1
	@echo "[DONE]  senz: sz.collector.v1 ... ok"

sz.collector.run: SENTINEZ_OUT ?= collector
sz.collector.run: sz.collector.build
	@$(SZCOLLECTOR_DIR)/v1/bin/$(SENTINEZ_OUT) --env_file=$(SZCOLLECTOR_DIR)/v1/.env

sz.collector.image.build: TAG ?= sentinez/collector
sz.collector.image.build:
	@docker buildx build -f $(SZCOLLECTOR_DIR)/v1/Dockerfile -t $(TAG):latest $(SZCOLLECTOR_DIR)/../..
