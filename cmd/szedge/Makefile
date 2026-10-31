# Targets for cmd/szedge; included by the root Makefile.
# Override the binary name with SENTINEZ_OUT and the image with TAG; SCALE=1
# builds the enterprise features of scale/ in (hack/make/scale.mk).
SZEDGE_DIR := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))

.PHONY: sz.edge.build sz.edge.run sz.edge.image.build

sz.edge.build: SENTINEZ_OUT ?= edge
sz.edge.build:
	@go build -tags "$(SZ_GOTAGS)" -ldflags="-s -w" -o $(SZEDGE_DIR)/v1/bin/$(SENTINEZ_OUT) $(SZEDGE_DIR)/v1
	@echo "[DONE]  senz: sz.edge.v1 ... ok"

sz.edge.run: SENTINEZ_OUT ?= edge
sz.edge.run: sz.edge.build
	@$(SZEDGE_DIR)/v1/bin/$(SENTINEZ_OUT) \
		--cert_file=$(SZEDGE_DIR)/v1/is.s6z.io.vn.cert \
		--cert_key_file=$(SZEDGE_DIR)/v1/is.s6z.io.vn.key \
		--proxy_config=$(SZEDGE_DIR)/v1/proxy.yaml \
		--env_file=$(SZEDGE_DIR)/v1/.env \
		--engine=quic

sz.edge.image.build: TAG ?= sentinez/edge
sz.edge.image.build:
	@docker buildx build --build-arg SZ_GOTAGS="$(SZ_GOTAGS)" \
		-f $(SZEDGE_DIR)/v1/Dockerfile -t $(TAG):latest $(SZEDGE_DIR)/../..
