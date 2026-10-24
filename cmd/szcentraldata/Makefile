# Targets for cmd/szcentraldata; included by the root Makefile.
# Override the binary name with SENTINEZ_OUT and the image with TAG.
SZCENTRALDATA_DIR := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))

.PHONY: sz.centraldata.build sz.centraldata.image.build

sz.centraldata.build: SENTINEZ_OUT ?= centraldata
sz.centraldata.build:
	@go build -ldflags="-s -w" -o $(SZCENTRALDATA_DIR)/v1/bin/$(SENTINEZ_OUT) $(SZCENTRALDATA_DIR)/v1
	@echo "[DONE]  senz: sz.centraldata.v1 ... ok"

sz.centraldata.image.build: TAG ?= sentinez/centraldata
sz.centraldata.image.build:
	@docker buildx build -f $(SZCENTRALDATA_DIR)/v1/Dockerfile -t $(TAG):latest $(SZCENTRALDATA_DIR)/../..
