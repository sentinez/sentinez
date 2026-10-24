# Targets for cmd/szrealtime; included by the root Makefile.
# Override the binary name with SENTINEZ_OUT.
SZREALTIME_DIR := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))

.PHONY: sz.realtime.build sz.realtime.run

sz.realtime.build: SENTINEZ_OUT ?= realtime
sz.realtime.build:
	@go build -ldflags="-s -w" -o $(SZREALTIME_DIR)/bin/$(SENTINEZ_OUT) $(SZREALTIME_DIR)
	@echo "[DONE]  senz: sz.realtime ... ok"

sz.realtime.run: SENTINEZ_OUT ?= realtime
sz.realtime.run: sz.realtime.build
	@$(SZREALTIME_DIR)/bin/$(SENTINEZ_OUT) --env_file=$(SZREALTIME_DIR)/.env
