# Targets for cmd/szdataplane; included by the root Makefile.
# Override the binary name with SENTINEZ_OUT.
SZDATAPLANE_DIR := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))

.PHONY: sz.dataplane.build sz.dataplane.run

sz.dataplane.build: SENTINEZ_OUT ?= dataplane
sz.dataplane.build:
	@go build -ldflags="-s -w" -o $(SZDATAPLANE_DIR)/v1/bin/$(SENTINEZ_OUT) $(SZDATAPLANE_DIR)/v1
	@echo "[DONE]  senz: sz.dataplane ... ok"

# The dataplane attaches eBPF/XDP programs inside the "gateway" netns.
sz.dataplane.run: SENTINEZ_OUT ?= dataplane
sz.dataplane.run: sz.dataplane.build
	sudo ip netns exec gateway $(SZDATAPLANE_DIR)/v1/bin/$(SENTINEZ_OUT) \
		--env_file=$(SZDATAPLANE_DIR)/v1/.env
